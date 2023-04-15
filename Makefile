.DEFAULT_GOAL=build

build:
	docker build -t caller:latest .

run-client:
	go run cmd/client/main.go plan.yaml

run-server: build
	docker run --rm -p 8080:8080 caller:latest --listen :8080

shell: build
	docker run --rm -it --entrypoint /bin/bash caller:latest

shellb: 
	docker build --target pre-build -t caller:pre-build . && \
	docker run --rm -it --entrypoint /bin/bash caller:pre-build
