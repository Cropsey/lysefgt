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
  0x00433f13 0x00433980 runtime.findrunnable()             /usr/local/go/src/runtime/proc.go:2525
  0x00435119 0x00434ee0 runtime.schedule()                 /usr/local/go/src/runtime/proc.go:3111
  0x0043566d 0x00435520 runtime.park_m()                   /usr/local/go/src/runtime/proc.go:3314
  0x00452d83 0x00452d40 runtime.mcall()                    /usr/local/go/src/runtime/asm_amd64.s:402
  0x00451f4e 0x00451e20 time.Sleep()                       /usr/local/go/src/runtime/time.go:177
  0x00457485 0x00457460 main.easyToFindFunctionName()      /root/lysefgt/code/sample_app/test_bin.go:10
  0x004574f7 0x004574e0 main.main()                        /root/lysefgt/code/sample_app/test_bin.go:18
  0x0042ef12 0x0042ed00 runtime.main()                     /usr/local/go/src/runtime/proc.go:145
```

Even this final iteration of the profiler, you can still stop the execution by
```
# ctrl+c
```{{exec interrupt}}

