# Step 2: Meet Your Sample App
Before we continue on our profiler building journey, let's take a moment to get familiar with our example application `sample_app`{{}}.
This humble little piece of software will serve as our experimental subject as we explore ELF, DWARF, and the magic of eBPF.

### Browsing the Source Code

To explore the code, you can either use the killercoda built-in editor theia conveniently available for you in the left-most tab `Editor`{{}},
or you can take a peak directly on [github.com](https://github.com/Cropsey/lysefgt/tree/main/code/sample_app).

```
cd /root/lysefgt/code/sample_app
```{{exec}}

It is just a single source file, with one `main`{{}} function as program entry executing in endless loop calling two other functions, 
`easyToFindFunctionName`{{}} and `alsoEasyToFindFunctionName`{{}}.

The `Makefile`{{}} instructs Go compiler to not perform any build optimizations and keeps the default behaviour of not stripping debug
symbols from the binary.
```
go build -gcflags '-l' -o testbin ./test_bin.go
```
While it is technically possible to a certain degree to profile optimized binaries without debug symbols,
it's much more convenient to at least keep the debug symbols for the tradeoff of a bigger binary. Turning off build optimizations
can have negative performance effect, but it introduces additional challanges for the profiler and `sample_app` doesn't care about
performance.

### Reading the Binary `sample_app`

In order to be able to write a profiler, it's useful to be familiar with a binary format of the executables. Go language has very
convenient toolchain to help us dive into the compiled binary.

```
go tool objdump -s '^main' ./testbin
go tool objdump ./testbin | wc -l
```{{exec}}

Given how small our codebase is, you are probably surprised by the size of `objdump`{{}}. As mentioned before, Go has a runtime 
which is bundled in every binary for them to be self-contained. You can tell `objdump`{{}} to display only symbols matching a
regular expression, that is exactly what the first command does with `-s '^main'`{{}}. If you want to see all ~80.000 lines, you
can drop the `-s`{{}} flag.

We will focus on the last three `TEXT`{{}} sections that are relevant to the codebase of `sample_app`{{}}:

```
TEXT main.easyToFindFunctionName(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:10        0x457460                493b6610                CMPQ 0x10(R14), SP
  ...
  test_bin.go:18        0x47e021                c3                      RET
  ...

TEXT main.alsoEasyToFindFunctionName(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:20        0x47e040                493b6610                CMPQ 0x10(R14), SP
  ...
  test_bin.go:28        0x47e0a1                c3                      RET
  ...

TEXT main.main(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:30        0x47e0c0                493b6610                CMPQ 0x10(R14), S
  ...
  test_bin.go:32        0x47e0d2                e8e9feffff              CALL main.easyToFindFunctionName(SB)
  test_bin.go:33        0x47e0d7                e864ffffff              CALL main.alsoEasyToFindFunctionName(SB)
  ...
```

The `TEXT`{{}} section marks a beginning of particular symbol, for example following line:
```
TEXT main.easyToFindFunctionName(SB) /root/lysefgt/code/sample_app/test_bin.go
```
This means that symbol `main.easyToFindFunctionName`{{}} which refers to `easyToFindFunctionName()`{{}} function in
`main`{{}} package can be found in `/root/lysefgt/code/sample_app/test_bin.go`{{}} source file.

```
  test_bin.go:10        0x457460                493b6610                CMPQ 0x10(R14), SP
```
This tells us an assembly language instruction `CMPQ 0x10(R14), SP`{{}} is at memory address `0x457460`{{}} in hexadecimal,
which corresponds to line `10`{{}} of source code file `test_bin.go`{{}}. The third column is the machine code for the instruction.

You may notice there are multiple instructions mapped to the same line of the source code file. While by some, Go is considered
to be a rather verbose programming language, it still is significantly more compact than assembly.

The `objdump`{{}} combines both ELF and DWARF into an output that is easier to understand and more than sufficient for the
scope of the workshop. But if you are interested in seeing ELF and DWARF much closer to the raw representation, you can use
following commands, but beware, there will be a lot of lines.

ELF:
```
readelf -a ./testbin > elfdump 2>&1
cat elfdump | wc -l
# elf output stored in a file ./elfdump
```{{exec}}

DWRF:
```
readelf -w ./testbin > dwarfdump 2>&1
cat dwarfdump | wc -l
# dwarf output stored in a file ./dwarfdump
```{{exec}}
