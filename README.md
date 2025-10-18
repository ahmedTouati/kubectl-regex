# kubectl-regex-match

This repository implements a [kubectl](https://github.com/kubernetes/kubectl) plugin for get/delete kubernetes resources by regex patterns.

## Features
🔍 Get resources by regex (Pods, Deployments, Services, ConfigMaps, CRDs, …)

🗑 Delete resources by regex with confirmation

🏷 Works with any resource supported by your cluster’s API

🌐 Supports --namespace, -n, --context, --kubeconfig, --all-namespaces like native kubectl

⚡ Powered by[ Go’s RE2 regex engine](https://github.com/google/re2)
 (safe, no catastrophic backtracking)

 ### 💡 Tired of long shell pipelines just to clean up resources?
Instead of typing something like to delete all pods that contain "app" in current namespace:

```bash
kubectl get pods | grep "app" | awk '{print $1}' | xargs kubectl delete pods
```
You can simply run:
```bash
kubectl regex-match delete pods "app"
```

## Install (local build)
```bash
git clone https://github.com/ahmedTouati/kubectl-regex-match.git
cd kubectl-regex-match
go build -o kubectl-regex-match cmd/kubectl-regex-match.go
mv kubectl-regex-match /usr/local/bin/
```

Check plugin discovery:
```bash
kubectl plugin list
```

## 🚀 Usage

Get resources
```bash
# Get pods starting with "nginx-"
kubectl regex-match get pods "^nginx-"

# Get services ending with "-svc"
kubectl regex-match get services ".*-svc"

# Get deployments containing "web" in namespace "foo"
kubectl regex-match get deployments "web" -n "foo"
```

Delete resources
```bash
# Delete all configmaps containing "app" in namespace foo
kubectl regex-match delete configmaps "app" -n foo

# Delete all deployments whose names start with "test-" in the default namespace, without asking for confirmation (use with caution)
kubectl regex-match delete deployments "^test-" --yes
```

All namespaces
```bash
kubectl regex-match get pods "nginx" -A
```

## ⚙️ Regex syntax

Uses [Go’s built-in regexp](https://github.com/google/re2)

## 📄 License

Apache 2.0 License.