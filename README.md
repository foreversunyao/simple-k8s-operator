# sam-operator
- based on operator-sdk(v1.26.0)
[refer](https://dev.to/austincunningham/build-a-kubernetes-operator-1g08)


# run on kind
```
kube-system          pod/etcd-sam-kind-1-control-plane                      1/1     Running   2 (154m ago)   177m
kube-system          pod/kube-apiserver-sam-kind-1-control-plane            1/1     Running   2 (154m ago)   177m
kube-system          pod/kube-controller-manager-sam-kind-1-control-plane   1/1     Running   2 (154m ago)   177m
kube-system          pod/kube-scheduler-sam-kind-1-control-plane            1/1     Running   2 (154m ago)   177m
```
# run on local
```
test -s /Users/samuelsun/20230331_tmp/sam-operator/bin/controller-gen || GOBIN=/Users/samuelsun/20230331_tmp/sam-operator/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0
/Users/samuelsun/20230331_tmp/sam-operator/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
test -s /Users/samuelsun/20230331_tmp/sam-operator/bin/kustomize || { curl -Ss "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- 3.8.7 /Users/samuelsun/20230331_tmp/sam-operator/bin; }
/Users/samuelsun/20230331_tmp/sam-operator/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/sams.sam.io unchanged
/Users/samuelsun/20230331_tmp/sam-operator/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./main.go
1.680257828361444e+09	INFO	controller-runtime.metrics	Metrics server is starting to listen	{"addr": ":8080"}
1.680257828361702e+09	INFO	setup	starting manager
1.680257828361969e+09	INFO	Starting server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
1.68025782836197e+09	INFO	Starting server	{"kind": "health probe", "addr": "[::]:8081"}
1.680257828362087e+09	INFO	Starting EventSource	{"controller": "sam", "controllerGroup": "sam.io", "controllerKind": "Sam", "source": "kind source: *v1alpha1.Sam"}
1.680257828362108e+09	INFO	Starting Controller	{"controller": "sam", "controllerGroup": "sam.io", "controllerKind": "Sam"}
1.6802578284633431e+09	INFO	Starting workers	{"controller": "sam", "controllerGroup": "sam.io", "controllerKind": "Sam", "worker count": 1}
^C1.6802578989363918e+09	INFO	Stopping and waiting for non leader election runnables
1.6802578989365098e+09	INFO	Stopping and waiting for leader election runnables
1.6802578989365451e+09	INFO	Shutdown signal received, waiting for all workers to finish	{"controller": "sam", "controllerGroup": "sam.io", "controllerKind": "Sam"}
1.680257898936623e+09	INFO	All workers finished	{"controller": "sam", "controllerGroup": "sam.io", "controllerKind": "Sam"}
1.6802578989366558e+09	INFO	Stopping and waiting for caches
1.680257898936754e+09	INFO	Stopping and waiting for webhooks
1.6802578989368389e+09	INFO	Wait completed, proceeding to shutdown the manager
```

