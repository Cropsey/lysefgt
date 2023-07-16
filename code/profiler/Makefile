all: profiler

profiler: main.go parser.go bpf_bpfel_x86.o
	go build -o profiler ./main.go ./parser.go ./bpf_bpfel_x86.go

# compile eBPF program using cillium/ebpf library
bpf_bpfel_x86.o: main.go perf.c
	CPATH=../headers go generate ./...

.PHONY: clean
clean:
	rm -rf profiler bpf_bpfel_x86.go bpf_bpfel_x86.o

.PHONY: docker-build
docker-build: profiler
	docker build . -t registry.kube-system.svc:5000/profiler:latest
	docker push registry.kube-system.svc:5000/profiler:latest

.PHONY: deploy
deploy: docker-build
	kubectl apply -f daemonset.yaml
