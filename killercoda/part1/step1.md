# Step 1: Setup, Build, Run, and Profile
We will start rather unusually, from the very end reaching our desired goal **\*immediatelly\***.

In this step, we'll build everything and ensure it all works as expected. It's the following steps where we will need
to roll up our sleeves and flex those programmer muscles, but don't worry, this guide will accompany you through the whole process.

The background script should finish installing all the necessary tools any moment now, it executed one command
but downloading the packages occassionally can take some time, especially over slower connection.
```
apt-get install clang llvm
```

### Working repository
We will need to clone the working repository, build and run the apps.
```
git clone https://github.com/Cropsey/lysefgt.git
```{{exec}}

The first application is called `sample_app`{{}}, this is the program we will profile. It's trivial on purpose
so we have easier time diving into the code and stack traces later. Let's just build it and run it now so
we can profile it later.
```
cd /root/lysefgt/code/sample_app
make
./testbin
```{{exec}}

As it has taken over the terminal, we'll keep it running there for the whole workshop. Let's open a new terminal tab inside killercoda
and continue there with building the profiler `perf_hacking`{{}} and profile our `sample_app`{{}}.
```
cd /root/lysefgt/code/perf_hacking
make bpfperf
```{{exec}}
We need to get the PID of the `sample_app`{{}} and pass it to the `perf_hacking`{{}} so it knows which process to profile.
```
pid=$(ps aux | awk '/testbin/{print($2)}')
./bpfperf -v 2 -pid $pid
```{{exec}}

### First Profiler Stack Traces
You should start to see sample stack traces, similar to this one:
```
bin[testbin] pid[24379]
  ADDRESS    PC         SYMBOL                             FILE:LINE
  ---------  ---------  ---------------------------------  ------------------------------------
  0x0047e022 0x0047dfc0 main.easyToFindFunctionName()      /root/lysefgt/code/sample_app/test_bin.go:10
  0x0047e0d7 0x0047e0c0 main.main()                        /root/lysefgt/code/sample_app/test_bin.go:30
  0x004324d2 0x004322c0 runtime.main()                     /usr/local/go/src/runtime/proc.go:14
```
You can stop the profiler execution by 
```
# ctrl+c
```{{exec interrupt}}

The stack traces read from top to bottom, meaning `main.easyToFindFunctionName`{{}} was the last function called and `runtime.main`{{}}
was the first one. We are profiling application written in Go and because Go has a runtime, there are going
to be some functions outside of the scope of our interest. But there are two functions that relate
directly to our `sample_app`{{}} - `main.main`{{}} and `main.easyToFindFunctionName`{{}}.

In the next step, we will take a closer look at the `sample_app`{{}}, it's code, some ELF and some DWARF.
