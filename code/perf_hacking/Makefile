all: bpfperf

bpfperf: main.go bpf_bpfel_x86.o
	go build -o bpfperf ./main.go ./bpf_bpfel_x86.go

# compile eBPF program using cillium/ebpf library
bpf_bpfel_x86.o: main.go perf.c
	CPATH=../headers go generate ./...

.PHONY: clean
clean:
	rm -rf bpfperf bpf_bpfel_x86.go bpf_bpfel_x86.o
