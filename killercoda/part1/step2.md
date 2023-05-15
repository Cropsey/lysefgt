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

In order to be able to write a profiler, it's useful to be familiar with a binary format of the applications. Go language has very
convenient toolchain to help us dive into the compiled binary.

```
go tool objdump ./testbin | wc -l
go tool objdump -s '^main' ./testbin
```{{exec}}

Given how small our codebase is, you are probably surprised by the size of `objdump`{{}}. As mentioned before, Go has a runtime 
which is bundled in every binary for them to be self-contained. You can tell `objdump`{{}} to display only symbols matching a
regular expression, that is exactly what the second command does with `-s '^main'`{{}}. If you want to see all ~80.000 lines, you
can drop the `-s`{{}} flag.

We will focus on the last three `TEXT`{{}} sections that are relevant to the codebase of `sample_app`{{}}:

```
TEXT main.easyToFindFunctionName(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:10        0x457460                493b6610                CMPQ 0x10(R14), SP
  test_bin.go:10        0x457464                7629                    JBE 0x45748f
  test_bin.go:10        0x457466                4883ec10                SUBQ $0x10, SP
  test_bin.go:10        0x45746a                48896c2408              MOVQ BP, 0x8(SP)
  test_bin.go:10        0x45746f                488d6c2408              LEAQ 0x8(SP), BP
  test_bin.go:11        0x457474                48b800f2052a01000000    MOVQ $0x12a05f200, AX
  test_bin.go:11        0x45747e                6690                    NOPW
  test_bin.go:11        0x457480                e89ba9ffff              CALL time.Sleep(SB)
  test_bin.go:12        0x457485                488b6c2408              MOVQ 0x8(SP), BP
  test_bin.go:12        0x45748a                4883c410                ADDQ $0x10, SP
  test_bin.go:12        0x45748e                c3                      RET
  test_bin.go:10        0x45748f                e86cbaffff              CALL runtime.morestack_noctxt.abi0(SB)
  test_bin.go:10        0x457494                ebca                    JMP main.easyToFindFunctionName(SB)

TEXT main.alsoEasyToFindFunctionName(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:14        0x4574a0                493b6610                CMPQ 0x10(R14), SP
  test_bin.go:14        0x4574a4                7629                    JBE 0x4574cf
  test_bin.go:14        0x4574a6                4883ec10                SUBQ $0x10, SP
  test_bin.go:14        0x4574aa                48896c2408              MOVQ BP, 0x8(SP)
  test_bin.go:14        0x4574af                488d6c2408              LEAQ 0x8(SP), BP
  test_bin.go:15        0x4574b4                48b800f2052a01000000    MOVQ $0x12a05f200, AX
  test_bin.go:15        0x4574be                6690                    NOPW
  test_bin.go:15        0x4574c0                e85ba9ffff              CALL time.Sleep(SB)
  test_bin.go:16        0x4574c5                488b6c2408              MOVQ 0x8(SP), BP
  test_bin.go:16        0x4574ca                4883c410                ADDQ $0x10, SP
  test_bin.go:16        0x4574ce                c3                      RET
  test_bin.go:14        0x4574cf                e82cbaffff              CALL runtime.morestack_noctxt.abi0(SB)
  test_bin.go:14        0x4574d4                ebca                    JMP main.alsoEasyToFindFunctionName(SB)

TEXT main.main(SB) /root/lysefgt/code/sample_app/test_bin.go
  test_bin.go:18        0x4574e0                493b6610                CMPQ 0x10(R14), SP
  test_bin.go:18        0x4574e4                7618                    JBE 0x4574fe
  test_bin.go:18        0x4574e6                4883ec08                SUBQ $0x8, SP
  test_bin.go:18        0x4574ea                48892c24                MOVQ BP, 0(SP)
  test_bin.go:18        0x4574ee                488d2c24                LEAQ 0(SP), BP
  test_bin.go:20        0x4574f2                e869ffffff              CALL main.easyToFindFunctionName(SB)
  test_bin.go:21        0x4574f7                e8a4ffffff              CALL main.alsoEasyToFindFunctionName(SB)
  test_bin.go:20        0x4574fc                ebf4                    JMP 0x4574f2
  test_bin.go:18        0x4574fe                6690                    NOPW
  test_bin.go:18        0x457500                e8fbb9ffff              CALL runtime.morestack_noctxt.abi0(SB)
  test_bin.go:18        0x457505                ebd9                    JMP main.main(SB)
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
