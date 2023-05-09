### Step 6: Enhancing Stack Traces with `"debug/dwarf"`
If you thought things were getting exciting, just wait till we bring in `"debug/dwarf"`! In this step, we'll enhance our stack traces further, coupling function names with their respective source files and concrete line numbers.
By the end of this step, your stack traces will contain sufficient amount of information, providing the level of insight you need to truly understand your application's performance.

The final set of enhancements are available at https://github.com/Cropsey/lysefgt/pull/6.

```
cd /root/lysefgt/code/perf_hacking
git fetch origin perf_hacking_4:perf_hacking_4
git merge perf_hacking_4
pid=$(ps aux | awk '/testbin/{print($2)}')
make clean
make bpfperf
./bpfperf -pid $pid
```{{exec}}

Now the stack traces should match what we saw in the beginning of the workshop. The source code file path as well as line number come from DWARF.
```
bin[testbin] pid[47794]
  ADDRESS    PC         SYMBOL                             FILE:LINE
  ---------  ---------  ---------------------------------  ------------------------------------
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()      /root/lysefgt/code/sample_app/test_bin.go:10
  0x0047e0d7 0x0047e0c0 main.main()                        /root/lysefgt/code/sample_app/test_bin.go:30
  0x004324d2 0x004322c0 runtime.main()                     /usr/local/go/src/runtime/proc.go:145
```

Even this iteration of the profiler, you can still stop the execution by
```
# ctrl+c
```{{exec interrupt}}

