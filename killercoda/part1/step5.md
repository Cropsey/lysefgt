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
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()    
  0x0047e0d7 0x0047e0c0 main.main()                      
  0x004324d2 0x004322c0 runtime.main()                   
  0x0045a901 0x0045a900 runtime.goexit.abi0()
```

As usual, you can stop the profiler execution again by
```
# ctrl+c
```{{exec interrupt}}
