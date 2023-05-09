### Step 4: Stacking Up Your Profiler
With the basics secured, it's time to enhance our profiler. In this step, we'll learn how to make the profiler useful with stack traces, powered by the eBPF helper function `bpf_get_stack`.
Let's take a look at first set of improvements in [PR#14](https://github.com/Cropsey/lysefgt/pull/14). Don't forget to swap back to terminal in **tab 2**.
```
cd /root/lysefgt/code/profiler
git fetch origin profiler_2:profiler_2
git merge profiler_2
pid=$(ps aux | awk '/sample_app/{print($2)}')
make clean
make profiler
./profiler -pid $pid
```{{exec}}

Now the stack traces may start to resamble the desired output, here are the promised hexadecimal numbers:
```
bin[sample_app] pid[47794]
  ADDRESS
  ---------
  0x0047e022
  0x0047e0d7
  0x004324d2
  0x0045a901
```
It's just the addresses, no ELF, no DWARF. We re-implemented the `perf_event`{{}} eBPF program from [asm to C](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L48)
and started to leverage the [cilium/ebpf](https://github.com/cilium/ebpf) toolchain even more with automatic eBPF scaffold generation and program compilation 
[using `go:generate`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/main.go#L18).

The type of eBPF program remains `perf_event`{{}} but we changed [the name to `stacktrace`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L47).
We continue using [eBPF map `BPF_MAP_TYPE_PERF_EVENT_ARRAY`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L30) just  
[renamed to `events`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L31) and there is a [new eBPF map `stackdata_map`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L39)
for [getting the stack traces](https://github.com/Cropsey/lysefgt/blob/profiler_2/code/profiler/perf.c#L65) from kernel via [`bpf_get_stack()`{{}}  eBPF helper function](https://man7.org/linux/man-pages/man7/bpf-helpers.7.html).

To confirm this, we can use the `bpftool` again in **tab 3**:
```
bpftool prog
```{{exec}} 
to observe something similar to this:
```
277: perf_event  name perf_event_stacktrace  tag e303e8e6f5ff8394  gpl
        loaded_at 2023-06-14T09:49:48+0000  uid 0
        xlated 384B  jited 224B  memlock 4096B  map_ids 15,17,18
        btf_id 16
```
and we can look at the maps as well
```
bpftool map
```{{exec}}
expecting similar output to this
```
15: percpu_array  name stackdata_map  flags 0x0
        key 4B  value 192B  max_entries 1  memlock 4096B
        btf_id 14
18: perf_event_array  name events  flags 0x0
        key 4B  value 4B  max_entries 1  memlock 4096B
```

You can stop the profiler execution in **tab 2** again by
```
# ctrl+c
```{{exec interrupt}}

The eBPF program is pre-compiled this time and its ELF is much smaller than `sample_app`{{}} ELF.

```
readelf -a ./bpf_bpfel_x86.o
```{{exec}}

```
ELF Header:
  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00 
  Class:                             ELF64
...
  Machine:                           Linux BPF
...
  Entry point address:               0x0
```

In the next step, we will try to talk to Legolas to get some information from the ELFs.
