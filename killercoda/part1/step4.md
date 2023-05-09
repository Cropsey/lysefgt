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
  0x00457fc3
  0x004098e7
  0x00433bcc
  0x0043541c
  0x00436251
  0x0043676d
  0x00455d23
  0x00454ef5
  0x0045a785
  0x0045a7f7
  0x0042fbe7
  0x00456141
```

You can stop the profiler execution again by
```
# ctrl+c
```{{exec interrupt}}
