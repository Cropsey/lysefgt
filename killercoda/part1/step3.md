### Step 3: Baby Steps towards Profiling
Ready to get started with your first profiler? Let's not rush. It's important to grasp the fundamentals before we move onto the fun stuff.

Let's go back in time to a place where there is no `perf_hacking` profiler.
```
cd /root/lysefgt/code/perf_hacking
make clean
cd ..
git checkout perf_hacking_0
```{{exec}}

Because it would be difficult to write every line of the profiler in a 60 minute workshop, we will instead incrementaly develop by reviewing
and merging PRs. And let's start by reviewing the initial https://github.com/Cropsey/lysefgt/pull/3. It creates a basis that we will be
able to gradually build up.

```
git fetch origin perf_hacking_1:perf_hacking_1
git merge perf_hacking_1
cd /root/lysefgt/code/perf_hacking
pid=$(ps aux | awk '/testbin/{print($2)}')
make clean
make bpfperf
./bpfperf -pid $pid
```{{exec}}

You should get much simpler events from eBPF, not resembling the stack traces at all:
```
pid[47794]
  RECORD
  ---------
  {1 [123 0 0 0] 0}
```
You can stop the profiler execution again by
```
# ctrl+c
```{{exec interrupt}}
