### Step 3: Baby Steps towards Profiling
Ready to get started with your first profiler? Let's not rush. It's important to grasp the fundamentals before we move onto the fun stuff.

Let's go back in time to a place where there is no `profiler`{{}}.
```
cd /root/lysefgt/code/profiler
make clean
cd ..
git checkout profiler_0
```{{exec}}

Because it would be difficult to write every line of the profiler in a 60 minute workshop, we will instead incrementaly develop by reviewing
and merging PRs. And let's start by reviewing the initial [PR#13](https://github.com/Cropsey/lysefgt/pull/13). It creates a basis that we will be
able to gradually build up.

```
git fetch origin profiler_1:profiler_1
git merge profiler_1
cd /root/lysefgt/code/profiler
pid=$(pgrep sample_app)
make clean
make profiler
./profiler -pid $pid
```{{exec}}

You should get much simpler events from eBPF, not resembling the stack traces at all:
```
pid[47794]
  RECORD
  ---------
  {1 [123 0 0 0] 0}
```

The reason is, the foundation we are starting with uses a [`perf_event_open`{{}} syscall](https://man7.org/linux/man-pages/man2/perf_event_open.2.html) 
to instruct kernel to emit events of [`PERF_SAMPLE_RAW`{{}} type](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L91),
which means the format of the event and what it will contain is defined elsewhere. In our profiler, it's defined by [the eBPF program](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L56-L74)
to be just [simple `123`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L58).

Keep both applications running and let's take a look under the hood using [`bpftool`{{}}](https://github.com/libbpf/bpftool). You will need to
**open another terminal** because your first tab should be taken by the `sample_app`{{}} and the second tab by the `profiler`{{}}.

Now, in the **third tab**, try calling `bpftool`{{}} to inspect loaded eBPF programs:
```
bpftool prog
```{{exec}}
What you should see a few eBPF programs (many things in the Linux world use eBPF now) and among them our program of [type `perf_event`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L48)
with [name `my_perf_event_prog`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L47).
```
...
273: perf_event  name my_perf_event_p  tag b25036dae30edc28  gpl
        loaded_at 2023-06-14T09:45:23+0000  uid 0
        xlated 96B  jited 73B  memlock 4096B  map_ids 4
...
```
There is also eBPF map defined for the `my_perf_event_prog`{{}} to send valuable `123`{{}} messages to the userspace `profiler`{{}} with
[name `my_perf_array`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_1/code/profiler/main.go#L31).
```
bpftool map
```{{exec}}
```
4: perf_event_array  name my_perf_array  flags 0x0
        key 4B  value 4B  max_entries 1  memlock 4096B
```

You can stop the `profiler`{{}} execution in **tab 2** again by:
```
# ctrl+c
```{{exec interrupt}}

In the next step, we will print some hexadecimal numbers instead of `123`{{}} and see how much we remember about ELF.
