all: profiler

profiler: main.go parser.go bpf_bpfel_x86.o
	go build -o profiler ./main.go ./parser.go ./bpf_bpfel_x86.go

# compile eBPF program using cillium/ebpf library
bpf_bpfel_x86.o: main.go perf.c
	CPATH=../headers go generate ./...

.PHONY: clean
clean:
	rm -rf profiler bpf_bpfel_x86.go bpf_bpfel_x86.o
