### Step 4: Stacking Up Your Profiler
With the basics secured, it's time to enhance our profiler. In this step, we'll learn how to make the profiler useful with stack traces, powered by the eBPF helper function `bpf_get_stack`.
Let's take a look at first set of improvements https://github.com/Cropsey/lysefgt/pull/4.
```
cd /root/lysefgt/code/perf_hacking
git fetch origin perf_hacking_2:perf_hacking_2
git merge perf_hacking_2
pid=$(ps aux | awk '/testbin/{print($2)}')
make clean
make bpfperf
./bpfperf -pid $pid
```{{exec}}

Now the stack traces may start to resamble the desired output:
```
bin[testbin] pid[47794]
  ADDRESS
  ---------
  0x0047e022
  0x0047e0d7
  0x004324d2
  0x0045a901
```

You can stop the profiler execution again by
```
# ctrl+c
```{{exec interrupt}}
