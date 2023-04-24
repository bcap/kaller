IMAGE=bcap/caller
KIND_CLUSTER=caller
KIND_NAMESPACE=default

KIND=kind --name ${KIND_CLUSTER}
KUBECTL=kubectl --context kind-${KIND_CLUSTER} --namespace ${KIND_NAMESPACE}

.DEFAULT_GOAL=build

build:
	docker build -t ${IMAGE}:latest .

run-client-bare:
	go run cmd/client/main.go plan.yaml

run-server-bare:
	go run cmd/server/main.go $(args)

run-server: build
	docker run --rm -p 8080:8080 ${IMAGE}:latest --listen :8080 $(args)

shell: build
	docker run --rm -it --entrypoint /bin/bash ${IMAGE}:latest

shellb:
	docker build --target pre-build -t ${IMAGE}:pre-build . && \
	docker run --rm -it --entrypoint /bin/bash ${IMAGE}:pre-build

#
# local kubernetes dev/testing with kind
#

kind-cluster-create:
	${KIND} create cluster --config k8s/kind/cluster.yaml

kind-cluster-delete:
	${KIND} delete cluster

kind-load-image: build
	${KIND} load docker-image ${IMAGE}:latest

kind-undeploy:
	${KUBECTL} delete -f k8s/kind/caller.yaml

kind-deploy: kind-load-image kind-undeploy
	${KUBECTL} apply -f k8s/kind/caller.yaml

kind-tunnel:
	${KUBECTL} port-forward service/svc1 8080:80

kind-log-tail:
	stern --context kind-${KIND_CLUSTER} --namespace ${KIND_NAMESPACE} --since 10m --color=never --template '{{printf "%-21s %s\n" .PodName .Message}}' 'svc.*'

kind-htop-control:
	docker exec -it ${KIND_CLUSTER}-control-plane /bin/bash -c 'apt-get update && apt install -y htop && htop'

kind-htop-worker:
	docker exec -it ${KIND_CLUSTER}-worker /bin/bash -c 'apt-get update && apt install -y htop && htop'