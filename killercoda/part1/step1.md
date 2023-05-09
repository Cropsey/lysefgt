# Step 1: Setup, Build, Run, and Profile
We will start rather unusually, from the very end reaching our desired goal **\*immediatelly\***.

In this step, we'll build everything and ensure it all works as expected. It's the following steps where we will need
to roll up our sleeves and flex those programmer muscles, but don't worry, this guide will accompany you through the whole process.

The background script should finish installing all the necessary tools any moment now, downloading the packages occassionally can take
some time, especially over slower connection. The prerequisites are [`bpftool`{{}}](https://github.com/libbpf/bpftool), [`llvm`{{}}](https://llvm.org/) and [`clang`{{}}](https://clang.llvm.org/).

### Working repository
We will need to clone the working repository, build and run the apps.
```
git clone https://github.com/Cropsey/lysefgt.git -b code
```{{exec}}

The first application is called `sample_app`{{}}, this is the program we will profile. It's trivial on purpose
so we have easier time diving into the code and stack traces later. Let's just build it and run it now so
we can profile it later.
```
cd /root/lysefgt/code/sample_app
make
./sample_app
```{{exec}}

As it has taken over the terminal, we'll keep it running there for the whole workshop. Let's open a **new terminal tab** inside killercoda
and continue there with building the `profiler`{{}} program and get stack traces for our `sample_app`{{}}.
```
cd /root/lysefgt/code/profiler
make
```{{exec}}
We need to get the PID of the `sample_app`{{}} and pass it to the `profiler`{{}} so it knows which process to profile.
```
pid=$(ps aux | awk '/sample_app/{print($2)}')
./profiler -v 2 -pid $pid
```{{exec}}

### First Profiler Stack Traces
You should start to see sample stack traces, similar to this one:
```
bin[sample_app] pid[24379]
  ADDRESS    PC         SYMBOL                             FILE:LINE
  ---------  ---------  ---------------------------------  ------------------------------------
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()      /root/lysefgt/code/sample_app/main.go:10
  0x0047e0d7 0x0047e0c0 main.main()                        /root/lysefgt/code/sample_app/main.go:30
  0x004324d2 0x004322c0 runtime.main()                     /usr/local/go/src/runtime/proc.go:14
  0x0045a901 0x0045a900 runtime.goexit.abi0()              :0
```
You can stop the profiler execution by 
```
# ctrl+c
```{{exec interrupt}}

Upon termination, the profiler should have printed some aggregated statistics:
```
AGGREGATED PERF EVENT SAMPLES:
  COUNT  SYMBOL                                 FILE:LINE
  -----  -------------------------------------  ------------------------------------
     80  runtime.goexit.abi0()                  :0
     80  main.main()                            /root/lysefgt/code/sample_app/main.go:30
     80  runtime.main()                         /usr/local/go/src/runtime/proc.go:145
     41  main.easyToFindFunctionName()          /root/lysefgt/code/sample_app/main.go:10
     39  main.alsoEasyToFindFunctionName()      /root/lysefgt/code/sample_app/main.go:20
```

The stack traces read from top to bottom, meaning `main.easyToFindFunctionName`{{}} was the last function called and `runtime.main`{{}}
was the first one. We are profiling an application written in Go and because Go has a runtime, there are going
to be some functions outside of the scope of our interest. But there are three functions that relate
directly to our `sample_app`{{}} - `main.main`{{}}, `main.easyToFindFunctionName`{{}} and `main.alsoEasyToFindFunctionName`{{}}.
The aggregated stats are sorted by the most frequently occurring symbol to the lowest, meaning we took 80 samples where
all samples contained `runtime.main()`{{}}, but just roughly half contained `main.easyToFindFunctionName()`{{}} and other half
contianed `main.alsoEasyToFindFunctionName()`{{}}.

In the next step, we will take a closer look at the `sample_app`{{}}, it's code, some ELF and some DWARF.
