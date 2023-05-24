all: bpfperf

# compile eBPF program using cillium/ebpf library 
bpf_bpfel_x86.o: main.go perf.c
	CPATH=../headers go generate ./...

# perf
bpfperf: main.go parser.go bpf_bpfel_x86.o
	go build -o bpfperf ./main.go ./bpf_bpfel_x86.go ./parser.go

.PHONY: clean
clean:
	rm -rf bpf_bpfel_x86.go bpf_bpfel_x86.o bpfperf
