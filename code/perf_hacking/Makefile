all: bpfperf

bpfperf: main.go
	go build -o bpfperf ./main.go

.PHONY: clean
clean:
	rm -rf bpfperf
