# Step 2: Turn `sample_app`{{}} Into Containerized Application for Kubernetes
In [part 1](https://killercoda.com/wozniakjan/course/killercoda/part1), we have built arguably not very useful application and now
we are going to throw in some docker and Kubernetes to allow it to perform its useless calculations elsewhere - on somebody's Kubernetes cluster.

There is actually not that much work, [the `Dockerfile`{{}}](https://github.com/Cropsey/lysefgt/blob/code2/code/sample_app/Dockerfile) is trivial,
it just [copies the binary to a container image](https://github.com/Cropsey/lysefgt/blob/code2/code/sample_app/Dockerfile#L3) and [sets a command
to be executed as the container entry point](https://github.com/Cropsey/lysefgt/blob/code2/code/sample_app/Dockerfile#L5). This killercoda Kubernetes
environment has an image registry deployed and exposed as [registry.kube-system.svc:5000](https://github.com/Cropsey/lysefgt/blob/code2/code/sample_app/Makefile#L15-L16)
so that is how we tag the image and where we push it. Also it's the address our [`Deployment`{{}} for `sample_app`{{}}](https://github.com/Cropsey/lysefgt/blob/code2/code/sample_app/deployment.yaml#L18)
references to pull the image and run the container.

You can take a look at the Kubernetes resources associated with our `sample_app`{{}}:
```
kubectl get deployment sample-app
kubectl get pod -o wide --selector=app=sample-app
```{{exec}}

As well as observe the logs to verify it indeed is performing (or at least claiming to perform) its useless calculations.
```
kubectl logs deployment/sample-app
```{{exec}}

Next we will enable `profiler`{{}} to see more than just a single process and use a very fancy word to describe that.
