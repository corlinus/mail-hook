.PHONY: build
build:
	go build .

.PHONY: clean
clean:
	rm -f smtp-hook

.PHONY: run
run:
	go run . -c tmp/config.yml

.PHONY: dockerbuild
dockerbuild:
		docker run --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp golang:1.20 go build -v

.PHONY: dockerbuild
dockerbuild:
	docker build -t smtp-hook:latest .

.PHONY: dockerrun
dockerrun:
	docker run -ti smtp-hook:latest
