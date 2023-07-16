### Step 6: Enhancing Stack Traces with `"debug/dwarf"`{{}}
If you thought things were getting exciting, just wait till we bring in `"debug/dwarf"`{{}}! In this step, we'll enhance our stack traces further, coupling function names with their respective source files and concrete line numbers.
By the end of this step, your stack traces will contain sufficient amount of information, providing the level of insight you need to truly understand your application's performance. The final set of stack trace enhancements are available at [PR#16](https://github.com/Cropsey/lysefgt/pull/16).

```
cd /root/lysefgt/code/profiler
git fetch origin profiler_4:profiler_4
git merge profiler_4
pid=$(pgrep sample_app)
make clean
make profiler
./profiler -pid $pid
```{{exec}}

Now the stack traces should match what we saw in the beginning of the workshop. The source code file path as well as line number come from DWARF.
```
bin[sample_app] pid[47794]
  ADDRESS    PC         SYMBOL                             FILE:LINE
  ---------  ---------  ---------------------------------  ------------------------------------
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()      /root/lysefgt/code/sample_app/main.go:10
  0x0047e0d7 0x0047e0c0 main.main()                        /root/lysefgt/code/sample_app/main.go:30
  0x004324d2 0x004322c0 runtime.main()                     /usr/local/go/src/runtime/proc.go:145
```
The visible change is the last column `FILE:LINE`{{}}. This is achieved by two packages ["debug/dwarf"](https://pkg.go.dev/debug/dwarf) and 
[dwarf reader from delve](https://github.com/go-delve/delve/tree/v1.20.2/pkg/dwarf/reader) to make our life of parsing DWARF a little easier.
If you haven't used it yet, [delve](https://github.com/go-delve/delve) is a great debugger, tailored to the specifics of Go language.

The algorithm to find the right [DWARF entry](https://github.com/Cropsey/lysefgt/blob/profiler_5/code/profiler/parser.go#L134-L180) should
be slightly less cryptic now given what we covered in the previous workshop steps and [the theory part of the workshop + the slides](https://github.com/Cropsey/lysefgt/)
 but just in case you are not one of the live attendees and you are trying this at home by yourself, here is more or less how this can be achieved:
```
1) read the DWARF part of the binary
2) find a CompileUnit that has an AttributeName matching the symbol from ELF
2.1) in this CompileUnit, find LineEntry matching DWARF LowPC to the Address
3) add this information to the output
```{{}}

I wouldn't expect any new discoveries from `bpftool` but here it is
```
bpftool prog
bpftool map
```{{exec}}

Even this iteration of the profiler, you can still stop the execution by
```
# ctrl+c
```{{exec interrupt}}

Coming next: more numbers and some counters!
