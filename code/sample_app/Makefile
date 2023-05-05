testbin: test_bin.go
	go build -gcflags '-l' -o testbin ./test_bin.go

dwarfdump: testbin
	readelf -w ./testbin  > dwarfdump ||:

clean:
	rm -rf testbin dwarfdump
