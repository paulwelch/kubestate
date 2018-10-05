# kubestate

Kubernetes State Metrics Utility

## Building kubestate

Tool Requirements
* go
* dep
* glide (required for downstream projects, such as github.com/kubicorn - hopefully will fall off the requirements list at some point)
* kubectl (required to run, not necessarily to build)

1. go get -u github.com/paulwelch/kubestate

2. go build

3. go install

## Running kubestate

```bash
» kubestate --help                                            paul@PaulsMacBook
NAME:
   kubestate - Show kubernetes state metrics

USAGE:
   kubestate [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  path to kubeconfig (default: "~/.kube/config")
   --raw, -r                 Show raw response data format
   --json, -j                Show JSON format
   --filter value, -f value  Metric filter to show (default: "*")
   --help, -h                show help
   --version, -v             print the version
   
```