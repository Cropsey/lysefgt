### Step 7: Aggregating results 

Unless you have a stack trace processor in your head, you may find the independent stack traces
little inconvenient for determining where does your application spend precious CPU time. That is why
[PR#17](https://github.com/Cropsey/lysefgt/pull/17) adds aggregated results. Now you can control verbosity
with `-v`{{}} flag and get either just the aggregated result at the end of execution or continue to
observe independent stack traces as they arrive.

```
cd /root/lysefgt/code/profiler
git fetch origin profiler_5:profiler_5
git merge profiler_5
pid=$(pgrep sample_app)
make clean
make profiler
./profiler -v 2 -pid $pid
```{{exec}}

In order to get the aggregated results, you need to terminate the profiler this time
```
# ctrl+c
```{{exec interrupt}}

The aggregated table should look similar to this one, depending on how long you allow the
profiler to run. The count tells how many times a particular symbol was present in a stack
trace, then the usual - symbol and source file + code line.
```
AGGREGATED PERF EVENT SAMPLES:
  COUNT  SYMBOL                                 FILE:LINE
  -----  -------------------------------------  ------------------------------------
    217  main.main()                            /root/lysefgt/code/sample_app/main.go:30
    217  runtime.main()                         /usr/local/go/src/runtime/proc.go:145
    111  main.alsoEasyToFindFunctionName()      /root/lysefgt/code/sample_app/main.go:20
    106  main.easyToFindFunctionName()          /root/lysefgt/code/sample_app/main.go:10
```
Because both `easyToFindFunctionName`{{}} as well as `alsoEasyToFindFunctionName`{{}} are
equivalent in computational complexity, statistically they should get sampled roughly same
amount of time.

You can try to change the `sample_app`{{}} code to see if you get different results. For example
making the `easyToFindFunctionName()`{{}} run tenth of the original time 
```go
func easyToFindFunctionName() {
	fmt.Println("easyToFindFunctionName")
	sum := 0
	for i := 0; i < x/10; i++ {
		for j := 0; j < x; j++ {
			sum += i + j
		}
	}
}
```
should also show in the aggregated results
```
AGGREGATED PERF EVENT SAMPLES:
  COUNT  SYMBOL                                 FILE:LINE
  -----  -------------------------------------  ------------------------------------
     52  main.main()                            /root/lysefgt/code/sample_app/main.go:30
     52  runtime.main()                         /usr/local/go/src/runtime/proc.go:145
     48  main.alsoEasyToFindFunctionName()      /root/lysefgt/code/sample_app/main.go:20
      4  main.easyToFindFunctionName()          /root/lysefgt/code/sample_app/main.go:10
```

And that concludes the part 1, onto the part 2!