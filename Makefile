.DEFAULT_GOAL=build

build:
	docker build -t caller:latest .

run-client-bare:
	go run cmd/client/main.go plan.yaml

run-server-bare:
	go run cmd/server/main.go $(args)

run-server: build
	docker run --rm -p 8080:8080 caller:latest --listen :8080 $(args)

shell: build
	docker run --rm -it --entrypoint /bin/bash caller:latest

shellb:
	docker build --target pre-build -t caller:pre-build . && \
	docker run --rm -it --entrypoint /bin/bash caller:pre-build
