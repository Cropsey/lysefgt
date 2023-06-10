all: argonsrv argoncli argonbin

argonsrv: main.go
	go build -gcflags '-l' -o argonsrv ./main.go

argoncli: cli.go
	go build -gcflags '-l' -o argoncli ./cli.go

argonbin: bin.go
	go build -gcflags '-l' -o argonbin ./bin.go

dwarfdump: argonsrv
	readelf -w ./argonsrv  > dwarfdump ||:

clean:
	rm -rf argonsrv dwarfdump
