# kubectl-regex

This repository implements a [kubectl](https://github.com/kubernetes/kubectl) plugin for get/delete kubernetes resources by regex patterns.

## Features
🔍 Get resources by regex (Pods, Deployments, Services, ConfigMaps, CRDs, …)

🗑 Delete resources by regex

🏷 Works with any resource supported by your cluster’s API

🌐 Supports --namespace, -n, --context, --kubeconfig, --all-namespaces like native kubectl

⚡ Powered by[ Go’s RE2 regex engine](https://github.com/google/re2)
 (safe, no catastrophic backtracking)

 ### 💡 Tired of long shell pipelines just to clean up resources?
Instead of typing something like:

```bash
kubectl get pods | grep "app" | awk '{print $1}' | xargs kubectl delete pods
```
You can simply run:
```bash
kubectl regex delete pods "app"
```

## Install (local build)
```bash
git clone https://github.com/ahmedTouati/kubectl-regex.git
cd kubectl-regex
go build -o kubectl-regex cmd/kubectl-regex.go
mv kubectl-regex /usr/local/bin/
```

Check plugin discovery:
```bash
kubectl plugin list
```

## 🚀 Usage

Get resources
```bash
# Get pods starting with "nginx-"
kubectl regex get pods "^nginx-"

# Get services ending with "-svc"
kubectl regex get services ".*-svc"

# Get deployments containing "web" in namespace "foo"
kubectl regex get deployments "web" -n "foo"
```

Delete resources
```bash
# Delete all configmaps containing "app" in namespace foo
kubectl regex delete configmaps "app" -n foo

# Delete deployments starting with "test-" in default namespace
kubectl regex delete deployments "^test-"
```

All namespaces
```bash
kubectl regex get pods "nginx" -A
```

## ⚙️ Regex syntax

Uses [Go’s built-in regexp](https://github.com/google/re2)

## 📄 License

Apache 2.0 License.