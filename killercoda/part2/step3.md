# Step 3: Omniscient `profiler`{{}}
In [part 1](https://killercoda.com/wozniakjan/course/killercoda/part1), we have built a single process profiler. Now is the right time to convert
that profiler into a capable profiling agent that will sample not just a single process but every process! (assuming it knows how and can understand it)

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

### Kubernetes Agent
Let's check-out [PR#18](https://github.com/Cropsey/lysefgt/pull/18) where the first steps towards a profiling agent are taken.
```
cd /root/lysefgt/code/profiler
git checkout kubernetes_0
make build
./profiler -v 2 -pid 0
```{{exec}}
We will first run it locally, leave it for a bit to see some stack traces and then kill it.
```
# ctrl+c
```{{exec interrupt}}

Looks somewhat operational, at least we did get some stack traces for various PIDs, so let's deploy it to our Kubernetes cluster as a `DaemonSet`{{}} and observe its actions.
```
make deploy
```{{exec}}
There are couple of noteworthy details about the profiler `DaemonSet`{{}}:
* the container is running as privileged
* it has `/sys/fs/bpf`{{}} mounted from the host
* probably obvious but for completion - is a `DaemonSet`{{}} so we can run the agent on each node easily
```
kubectl get daemonset profiler
```{{exec}}
```
kubectl get pod -o wide --selector=app=profiler
```{{exec}}
We didn't get very far, let's investigate the `CrashLoop`{{}} state.
```
kubectl logs daemonset/profiler
```{{exec}}
First obstacle is trivial to solve
```
map create: operation not permitted (MEMLOCK may be too low, consider rlimit.RemoveMemlock)
```
[`RLIMIT_MEMLOCK`{{}} is a parameter](https://man7.org/linux/man-pages/man2/getrlimit.2.html) used in Linux to limit 
the amount of physical memory that any one user process may lock into RAM.

The memory-locking capability is useful for time-sensitive or real-time applications that cannot tolerate the delays 
of swapping memory contents to and from disk. Locking memory ensures that the data the application needs is always readily 
available in RAM, avoiding potential performance penalties due to page faults. The eBPF programs and eBPF maps leverage this
capability for performance reasons and may need a good amount of locked memory.

### Removing `RLIMIT_MEMLOCK`{{}}
With `rlimit.RemoveMemLock()`{{}} taken care of business in [PR#19](https://github.com/Cropsey/lysefgt/pull/19), we can try and see what other
roadblocks Kubernetes has for us.
```
git checkout kubernetes_1
make deploy
kubectl rollout restart daemonset/profiler
```{{exec}}
The profiler is no longer crashing, good
```
kubectl get pod -o wide --selector=app=profiler
```{{exec}}
But is it successfully profiling?
```
kubectl logs daemonset/profiler
```{{exec}}
Looks like the `/proc` filesystem doesn't appear to understand the PIDs kernel is sending us, let's compare host and pod
```
ps aux | grep [s]ample_app
```{{exec}}
```
pid=$(pgrep sample_app)
ls -l /proc/$pid/exe
```
While host can see the `sample_app`{{}} pod among the processes, the profiler agent is completely oblivious to `sample_app's`{{}} existence.
```
kubectl exec daemonset/profiler -- ps aux
```{{exec}}
This is due to [kernel PID namespaces](https://man7.org/linux/man-pages/man7/pid_namespaces.7.html), a well known layer of isolation for Linux processes.

### Adding `hostPID`{{}}
This has yet another trivial solution, [PR#20](https://github.com/Cropsey/lysefgt/pull/20) adds [`hostPID`{{}} to the `DaemonSet` of the profiler](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#hosts-namespaces).
```
git checkout kubernetes_2
make deploy
```{{exec}}
Let's check the logs to see how far we advanced
```
kubectl logs daemonset/profiler
```{{exec}}
And we observe another discrepancy, looks like the path `/usr/bin/sample_app`{{}} defined in `/proc/${PID}/exe` doesn't exist, where are you Waldo?
```
find / -type f -name sample_app
```{{exec}}
Of course, the `sample_app`{{}} is in a pod and [containerd](https://containerd.io/) overlay filesystem plays an important role where each container has its `/`.
The paths in this case would all be in this form: `/run/containerd/io.containerd.runtime.v2.task/k8s.io/${CONTAINER_ID}/rootfs/${PATH_FROM_PROC}`{{}}.

### Overlay FS
The [PR#21](https://github.com/Cropsey/lysefgt/pull/21) addresses the path discrepancy, the `CONTAINER_ID`{{}} is not something the Linux Kernel understands,
but we can figure it out from pod's [`cgroup`{{}}](https://man7.org/linux/man-pages/man7/cgroups.7.html).
```
pid=$(pgrep sample_app)
cat /proc/$pid/cgroup
```{{exec}}
Let's try this version of profiler then
```
git checkout kubernetes_3
make deploy
```{{exec}}
And what do the logs say?
```
kubectl logs daemonset/profiler
```{{exec}}
We got this!

### Metrics
Last step here is [PR#22](https://github.com/Cropsey/lysefgt/pull/22) which exports the metrics in Prometheus format on endpoint `/metrics`{{}} port `2112`{{}}.
```
git checkout kubernetes_4
make deploy
```{{exec}}

Next we will write queries to one of the mighty Greek Titans, the God of forethought and crafty counsel, Prometheus!
