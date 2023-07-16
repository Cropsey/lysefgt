### Step 5: Cracking the Code with `"debug/elf"`{{}}
Have you ever wished you could decipher machine code as if it was the original programming language? Well, we're about to fulfill that wish. In this step, we'll employ the Go core package `"debug/elf"`{{}}
to extract human-readable function names from the addresses returned by `bpf_get_stack`{{}}.
Let's review the following set of improvements in [PR#15](https://github.com/Cropsey/lysefgt/pull/15). 
```
cd /root/lysefgt/code/profiler
git fetch origin profiler_3:profiler_3
git merge profiler_3
pid=$(pgrep sample_app)
make clean
make profiler
./profiler -pid $pid
```{{exec}}

The stack traces are even richer by containing the symbols from ELF symbol table.
```
bin[sample_app] pid[47794]
  ADDRESS    PC         SYMBOL                            
  ---------  ---------  --------------------------------- 
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()    
  0x0047e0d7 0x0047e0c0 main.main()                      
  0x004324d2 0x004322c0 runtime.main()                   
  0x0045a901 0x0045a900 runtime.goexit.abi0()
```
The profiler takes the PID and figures out the executable binary associated with that process [from `/proc/$PID/exe`{{}}](https://github.com/Cropsey/lysefgt/blob/profiler_3/code/profiler/parser.go#L32).
It reads the binary and [using "debug/elf"](https://pkg.go.dev/debug/elf) parses the [symbol table](https://github.com/Cropsey/lysefgt/blob/profiler_3/code/profiler/parser.go#L47-L52).
For each address in the particular stack trace, it then gets [the symbol](https://github.com/Cropsey/lysefgt/blob/profiler_3/code/profiler/parser.go#L60-L73)
and [enhances the output](https://github.com/Cropsey/lysefgt/blob/profiler_3/code/profiler/parser.go#L15).

Given we didn't touch the eBPF program this time, `bpftool`{{}} should give roughly the same output as last time
```
bpftool prog
bpftool map
```{{exec}}

As usual, you can stop the profiler execution again in **terminal 2** by
```
# ctrl+c
```{{exec interrupt}}

And what valuable treasures does Gimli and his DWARF colleagues hide? Let's see!
