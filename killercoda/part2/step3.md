# Step 3: Omniscient `profiler`{{}}
In [part 1](https://killercoda.com/wozniakjan/course/killercoda/part1), we have built a single process profiler. Now is the right time to convert
that profiler into a capable profiling agent that will sample not just a single process but every process (assuming it knows how and can understand it)!

Passing `-pid=0`{{}} or not passing one at all to the `profiler`{{}} was a bug. It would not be able to profile [process 0](https://unix.stackexchange.com/questions/83322/which-process-has-pid-0).
So we will give `-pid=0`{{}} brand new meaning. It will instruct the `profiler`{{}} to not profile just a single process but instead all processes.
But it's not that simple, because [calling `perf_event_open`{{}}](https://man7.org/linux/man-pages/man2/perf_event_open.2.html) for each CPU and each PID
is not supported
```
pid == -1 and cpu == -1
       This setting is invalid and will return an error.
```
In [part 1, we used single pid for any CPU](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/main.go#L116-L117)
```
pid > 0 and cpu == -1
       This measures the specified process/thread on any CPU.
```
And in [part 2, we will use any pid for concrete CPU]() in a loop [calling `perf_event_open` for each CPU with `-1` value for PID](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/main.go#L94)
```
pid == -1 and cpu >= 0
       This measures all processes/threads on the specified CPU.
       This requires CAP_PERFMON (since Linux 5.8) or
       CAP_SYS_ADMIN capability or a
       /proc/sys/kernel/perf_event_paranoid value of less than 1.
```

There are additional challenges with running `profiler`{{}} in Kubernetes. Kernel sees all processes, but depending on the cluster [CRI](https://kubernetes.io/docs/concepts/architecture/cri/),
[there might be `pid_namespaces` enabled](https://man7.org/linux/man-pages/man7/pid_namespaces.7.html) so `profiler`{{}} may have hard time
seeing the same processes and depending on the `CRI`{{}} filesystem for container images, paths to executables in `/proc`{{}} will be different.
Calling `syscalls`{{}} from inside of an image requires certain [capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html) and
access to certain paths on the host filesystem.

This Killercoda environment runs [containerd](https://containerd.io/) as CRI so here are few little heuristics to mitigate these challenges:
1) Run with [`hostPID` set to true](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/daemonset.yaml#L20). This bypasses
`pid_namespaces`{{}} so PIDs should match.
2) Figure out the profiled executable binary path in the [overlay filesystem](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/parser.go#L148-L158)
3) Allow loading eBPF programs and creating eBPF maps by [mounting host `/run`{{}} and `/sys/fs/bpf`{{}}](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/daemonset.yaml#L28-L32)
4) Because [ain't no body got time for](https://www.youtube.com/watch?v=6gLMSf4afzo) figuring out the right capabilities, [run as `privileged`{{}}](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/daemonset.yaml#L26)

The `profiler`{{}} statistics are also [newly captured](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/metrics.go#L14-L17) in the [Prometheus](https://prometheus.io/) format,
[exposed](https://github.com/Cropsey/lysefgt/blob/code2/code/profiler/daemonset.yaml#L16-L18C35) on the `/metrics`{{}} endpoint, port `2112`{{}}.

Let's take a look at the Kubernetes resources associated with our `profiler`{{}}:
```
kubectl get daemonset profiler 
kubectl get pod -o wide --selector=app=profiler
```{{exec}}

And you can take a look at the logs too 
```
kubectl logs daemonset/profiler
```{{exec}}

Next we will write queries to one of the mighty Greek Titans, the God of forethought and crafty counsel, Prometheus!
