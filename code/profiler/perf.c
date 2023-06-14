// +build ignore

// Include headers from ebpf/cilium
#include "common.h"
#include "bpf_tracing.h"

// License must be present for each eBPF program
char __license[] SEC("license") = "Dual MIT/GPL";

// Flag to obtain user space stack trace (kernel space stack trace flag is '0')
enum {
    BPF_F_USER_STACK        = (1ULL << 8),
};

// Max size of the stack trace, our traces are fairly small
#define MAX_STACK_RAWTP 10

// Event we want to send from the kernel eBPF program to the user space
struct event {
    u32 pid;                           // Kernel returned PID of the profiled process
    u32 kern_stack_size;               // Size of the kernel stack trace
    u32 user_stack_size;               // Size of the user space stack trace
    u64 kern_stack[MAX_STACK_RAWTP];   // Array to hold the kernel stack trace 
    u64 user_stack[MAX_STACK_RAWTP];   // Array to hold the user space stack trace
    char name[16];                     // Binary name, 16 chars is kernel defined (sched.h, TASK_COMM_LEN)
};

// eBPF map used for sending the events from kernel to user space
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

// In order to read the stack data, we need to use eBPF map for safe memory handling
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event);
} stackdata_map SEC(".maps");

// Force emitting struct event into the ELF for eBPF program so cilium/ebpf package can generate scaffold for us
const struct event *unused __attribute__((unused));

// ELF section name for the eBPF program so kernel knows how and where to load it 
// perf_event is one of the eBPF program types
// stacktrace is our defined name of the program
SEC("perf_event/stacktrace")
int perf_event_stacktrace(struct pt_regs *ctx) {
    int max_len = MAX_STACK_RAWTP * sizeof(u64);

    // Because stack is limitted to 512 bytes for eBPF programs, larger chunks of memory are allocated in
    // per-cpu array map. The key is always going to be 0 as we are not using the map for anything but
    // temporary memory for the stack traces
    u32 key = 0;
    struct event *event;
    event = bpf_map_lookup_elem(&stackdata_map, &key);
    if (!event) {
        return 0;
    }

    // Use eBPF helpers to get PID, kernel comm, and stack traces
    event->pid = bpf_get_current_pid_tgid();
    bpf_get_current_comm(event->name, sizeof(event->name));
    event->kern_stack_size = bpf_get_stack(ctx, event->kern_stack, max_len, 0);
    event->user_stack_size = bpf_get_stack(ctx, event->user_stack, max_len, BPF_F_USER_STACK);

    // Using eBPF helper to print tracing information for debugging purposes
    // on older kernels ABI allows passing up to 4 parameters
    // can be viewed in /sys/kernel/debug/tracing/trace_pipe
    bpf_printk("sending event: name(%s) kernel_size(%d)/userspace_size(%d)", event->name, event->kern_stack_size, event->user_stack_size);

    // Send the event with stack traces from kernel to user space
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event, sizeof(*event));

    return 0;
}

