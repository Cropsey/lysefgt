### Background setup for Ubuntu
We use Ubuntu VM here in this killercoda scenario. The background script installs
all the necessary tools for this particular workshop.
```
apt-get install clang llvm
```

#### Working repository
Let's clone the code
```
git clone https://github.com/Cropsey/lysefgt.git
cd lysefgt/code/sample_app
```{{exec}}

Build and run a sample application that we will later profile
```
make
./testbin
```{{exec}}

Open a new terminal tab inside killercoda and continue there with building a sample app
Build the profiler
```
cd lysefgt/code/perf_hacking
make bpfperf
```{{exec}}
Get the PID of the sample app and run the profiler
```
pid=$(ps aux | awk '/testbin/{print($2)}' )
./bpfperf -pid $pid
```{{exec}}
