sample_app: main.go
	go build -gcflags '-l' -o sample_app ./main.go

dwarfdump: sample_app
	readelf -w ./sample_app  > dwarfdump ||:

objdump: sample_app
	go tool objdump ./sample_app

clean:
	rm -rf sample_app dwarfdump

.PHONY: docker-build
docker-build: sample_app
	docker build . -t registry.kube-system.svc:5000/sample_app:latest
	docker push registry.kube-system.svc:5000/sample_app:latest

.PHONY: deploy
deploy: docker-build
	kubectl apply -f deployment.yaml
