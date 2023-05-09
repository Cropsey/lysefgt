# Step 1: Profiling Agent for Your Kubernetes Cluster
Even the part 2 will again start from the very end. Let's see if we can build all of our
applications, get them successfully deployed, operational, and providing good value.

* **Prometheus** is be getting deployed in the background by the setup script and should be reachable and exposed externally
* **Grafana** also deployed by the background setup script and should be reachable and exposed externally
* **Container registry** likewise getting setup but we don't need it to be publicly exposed, it's just for internal usage
* **Application `sample_app`{{}}** will mature to a containarized application and we will deploy it to our kubernetes cluster
* **Profiler `profiler`{{}}** will mature to containerized agent running on each kubernetes node and sampling selected applications every 100ms and feeding stack trace stats to prometheus as `profiler_aggregate`{{}} metric

Let's check out the code:
```
git clone https://github.com/Cropsey/lysefgt.git -b code2
```{{exec}}

Build and deploy the `sample_app`{{}}:
```
cd /root/lysefgt/code/sample_app
make deploy
```{{exec}}

And build and deploy the `profiler`{{}}:
```
cd /root/lysefgt/code/profiler
make deploy
```{{exec}}

You can open the [prometheus]({{TRAFFIC_HOST1_30080}}) window and try following query, it may take up to a minute
for the metrics to start showing in there.
```
sort_desc (profiler_aggregated{})
```{{}}

And also you can finally view some graphs in [grafana]({{TRAFFIC_HOST1_30081}})!

Next we will learn how to put it all on a Kubernetes cluster.
