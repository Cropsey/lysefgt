// +build ignore

#include "common.h"
#include "bpf_tracing.h"

char __license[] SEC("license") = "Dual MIT/GPL";

/* flags for both BPF_FUNC_get_stackid and BPF_FUNC_get_stack. */
enum {
    BPF_F_SKIP_FIELD_MASK    = 0xffULL,
    BPF_F_USER_STACK        = (1ULL << 8),
/* flags used by BPF_FUNC_get_stackid only. */
    BPF_F_FAST_STACK_CMP    = (1ULL << 9),
    BPF_F_REUSE_STACKID        = (1ULL << 10),
/* flags used by BPF_FUNC_get_stack only. */
    BPF_F_USER_BUILD_ID        = (1ULL << 11),
};

#define MAX_STACK_RAWTP 100
struct event {
    u32 pid;
    u32 kern_stack_size;
    u32 user_stack_size;
    u64 kern_stack[MAX_STACK_RAWTP];
    u64 user_stack[MAX_STACK_RAWTP];
    char name[16]; // hard limit in kernel (sched.h, TASK_COMM_LEN)
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event);
} stackdata_map SEC(".maps");

// Force emitting struct event into the ELF.
const struct event *unused __attribute__((unused));

SEC("perf_event/stacktrace")
int perf_event_stacktrace(struct pt_regs *ctx) {
    int max_len = MAX_STACK_RAWTP * sizeof(u64);

    u32 key = 0;
    struct event *event;

    // stack is limitted to 512 bytes, larger chunks of memory are allocated in BPF per-cpu array map
    event = bpf_map_lookup_elem(&stackdata_map, &key);
    if (!event) {
        return 0;
    }

    event->pid = bpf_get_current_pid_tgid();
    event->kern_stack_size = bpf_get_stack(ctx, event->kern_stack, max_len, 0);
    event->user_stack_size = bpf_get_stack(ctx, event->user_stack, max_len, BPF_F_USER_STACK);

    bpf_printk("event->name: %s", event->name);
    bpf_get_current_comm(event->name, sizeof(event->name));

    // can be viewed in /sys/kernel/debug/tracing/trace_pipe
    bpf_printk("sending event %d, %d/%d", sizeof(*event), event->kern_stack_size, event->user_stack_size);
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event, sizeof(*event));

    return 0;
}

