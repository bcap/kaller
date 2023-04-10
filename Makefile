.DEFAULT_GOAL=build

build:
	docker build -t caller:latest .

run: build
	docker run --rm caller:latest 

shell: build
	docker run --rm -it --entrypoint /bin/bash caller:latest 

shellb: 
	docker build --target pre-build -t caller:pre-build . && \
	docker run --rm -it --entrypoint /bin/bash caller:pre-build
