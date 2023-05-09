### Step 5: Cracking the Code with `"debug/elf"`
Have you ever wished you could decipher machine code as if it was the original programming language? Well, we're about to fulfill that wish. In this step, we'll employ the Go core package `"debug/elf"` to extract human-readable function names from the addresses returned by `bpf_get_stack`.
Let's review the following set of improvements https://github.com/Cropsey/lysefgt/pull/5. 
```
cd /root/lysefgt/code/perf_hacking
git fetch origin perf_hacking_3:perf_hacking_3
git merge perf_hacking_3
pid=$(ps aux | awk '/testbin/{print($2)}')
make clean
make bpfperf
./bpfperf -pid $pid
```{{exec}}

The stack traces are even richer by containing the symbols from ELF symbol table.
```
bin[testbin] pid[47794]
  ADDRESS    PC         SYMBOL
  ---------  ---------  ---------------------------------
  0x00457fc3 0x00457fa0 runtime.futex.abi0()
  0x004098e7 0x00409860 runtime.notesleep()
  0x00433bcc 0x00433b40 runtime.stopm()
  0x0043541c 0x00434960 runtime.findRunnable()
  0x00436251 0x004361a0 runtime.schedule()
  0x0043676d 0x00436640 runtime.park_m()
  0x00455d23 0x00455ce0 runtime.mcall()
  0x00454ef5 0x00454dc0 time.Sleep()
  0x0045a785 0x0045a760 main.easyToFindFunctionName()
  0x0045a7f7 0x0045a7e0 main.main()
  0x0042fbe7 0x0042f9e0 runtime.main()
  0x00456141 0x00456140 runtime.goexit.abi0()
```

As usual, you can stop the profiler execution again by
```
# ctrl+c
```{{exec interrupt}}
