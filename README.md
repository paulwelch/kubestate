# kubestate

Kubernetes State Metrics Utility

## Building kubestate

Tool Requirements and Tested Versions
* go v1.11
* dep v0.5.0-1
* glide 0.13.0-dev (required for downstream projects, such as github.com/kubicorn - hopefully will fall off the requirements list at some point)
* kubectl v1.12.0 (required to run, not necessarily to build)

1. go get -u github.com/paulwelch/kubestate

2. go build

3. go install

## Running kubestate

```bash
NAME:
   kubestate - Show kubernetes state metrics

USAGE:
   kubestate [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     get      Get metric
     top      Show top resource consumption by deployment
     watch    Watch metric
     list     List metric families
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value     path to config (default: "~/.kube/config")
   --namespace value, -n value  namespace to show (default is all namespaces) (default: "*")
   --help, -h                   show help
   --version, -v                print the version
```

####Notes
* Optional add-on API service kube-state-metrics is required.  See https://github.com/kubernetes/kube-state-metrics 
* By default, kubestate uses kubectl config for cluster connection and authentication. If you can run kubectl commands successfully, then kubestate should also be able to connect to the same cluster.

