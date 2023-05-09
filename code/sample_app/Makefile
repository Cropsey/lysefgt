sample_app: main.go
	go build -gcflags '-l' -o sample_app ./main.go

dwarfdump: sample_app
	readelf -w ./sample_app  > dwarfdump ||:

objdump: sample_app
	go tool objdump ./sample_app

clean:
	rm -rf sample_app dwarfdump
