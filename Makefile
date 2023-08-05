IMAGE=bcap/caller
KIND_CLUSTER=caller
KIND_NAMESPACE=default

KIND=kind --name ${KIND_CLUSTER}
KUBECTL_NO_NS=kubectl --context kind-${KIND_CLUSTER}
KUBECTL=${KUBECTL_NO_NS} --namespace ${KIND_NAMESPACE}
ISTIOCTL=istioctl --context kind-${KIND_CLUSTER}

.DEFAULT_GOAL=build

build:
	docker build -t ${IMAGE}:latest .

push: build
	docker push ${IMAGE}:latest

dive: build
	dive ${IMAGE}:latest

run-client-bare:
	go run cmd/client/main.go examples/plan.yaml

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

kind-use-istio:
	${ISTIOCTL} install --set profile=demo --skip-confirmation
	${KUBECTL_NO_NS} label namespace default istio-injection=enabled
	${KUBECTL_NO_NS} apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/kiali.yaml
	${KUBECTL_NO_NS} apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/prometheus.yaml
	${KUBECTL_NO_NS} apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/grafana.yaml
	${KUBECTL_NO_NS} apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/jaeger.yaml
	${KUBECTL_NO_NS} --namespace=istio-system wait --for=condition=available deployment kiali --timeout=60s
	${KUBECTL_NO_NS} --namespace=istio-system wait --for=condition=available deployment prometheus --timeout=60s
	${KUBECTL_NO_NS} --namespace=istio-system wait --for=condition=available deployment grafana --timeout=60s
	${KUBECTL_NO_NS} --namespace=istio-system wait --for=condition=available deployment jaeger --timeout=60s
	
kind-use-metallb:
	${KUBECTL_NO_NS} apply -f k8s/kind/loadbalancer-01.yaml
	${KUBECTL_NO_NS} --namespace=metallb-system wait deployment --for=condition=available --selector=component=controller --timeout=60s
	${KUBECTL_NO_NS} apply -f k8s/kind/loadbalancer-02.yaml

kind-fresh: kind-cluster-delete kind-cluster-create kind-use-istio kind-deploy kind-wait-pods-ready
	${KUBECTL_NO_NS} get pods -A

kind-load-image: build
	${KIND} load docker-image ${IMAGE}:latest

kind-undeploy:
	${KUBECTL} delete -f k8s/kind/caller.yaml

kind-deploy: kind-load-image
	${KUBECTL} apply -f k8s/kind/caller.yaml

kind-wait-pods-ready:
	${KUBECTL} wait pod --selector='app in (svc1,svc2,svc3)' --for=condition=ready --timeout=60s

kind-tunnel:
	${KUBECTL} port-forward service/svc1 8080:80

kind-log-tail:
	stern --context kind-${KIND_CLUSTER} --namespace ${KIND_NAMESPACE} --since 10m --color=never --template '{{printf "%-21s %s\n" .PodName .Message}}' '(client|svc).*'

kind-htop:
	docker exec -it ${KIND_CLUSTER}-control-plane /bin/bash -c 'apt-get update && apt install -y htop && htop'

kind-shell:
	${KUBECTL} run -it --rm --restart=Never --image alpine temporary-shell -- sh
