# kubestate

Kubernetes State Metrics Utility

The primary use for the kubernetes [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics/) add-on service is to provide cluster state metrics to monitoring systems like [Prometheus](https://prometheus.io/). But, it may sometimes be useful to get the same metrics in a script or directly from your command line. That's where kubestate comes in. It's a command line utility that calls the kube-state-metrics API, then shows interesting views of the metrics. You can also use it to get the raw data values in various formats that can be used by scripts or other utilities.

## Building kubestate

Tool Requirements and Tested Versions
* go v1.11
* glide 0.13.0-dev (required for downstream projects, such as github.com/kubicorn - hopefully will fall off the requirements list at some point)
* kubectl v1.12.0 (required to run, not necessarily to build)

1. go get -u github.com/paulwelch/kubestate

2. go build

3. go install

#### Notes
* Optional add-on API service kube-state-metrics is required.  See https://github.com/kubernetes/kube-state-metrics 
* By default, kubestate uses kubectl config for cluster connection and authentication. If you can run kubectl commands successfully, then kubestate should also be able to connect to the same cluster.

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

#### Examples
One interesting insight the kube-state-metrics service provides is requested and limits of CPU and memory resources. Here's the kubernetes [documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) and an example of using kubestate to show them by pod.

```bash
~ » kubestate top pods
Namespace     Pod                                       Container                CPU (Req / Lim) Memory  (Req / Lim) Node Load
istio-system  istio-pilot-7d6549448f-6hkvn              discovery                (500m / 0m)     (2048Mi / 0Mi)      wrk6 13%
kube-system   canal-7zvgk                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk2 3%
kube-system   canal-96q62                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk3 3%
kube-system   canal-p7zvn                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk6 3%
kube-system   canal-mx8md                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk1 3%
kube-system   canal-6p8zp                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk5 3%
kube-system   canal-9lmff                               calico-node              (250m / 0m)     (0Mi / 0Mi)         wrk4 3%
kube-system   kube-dns-7588d5b5f5-n4hqv                 dnsmasq                  (150m / 0m)     (20Mi / 0Mi)        wrk4 2%
kube-system   kube-dns-7588d5b5f5-8gzbp                 dnsmasq                  (150m / 0m)     (20Mi / 0Mi)        wrk4 2%
kube-system   kube-state-metrics-679d95df65-bp7cp       kube-state-metrics       (106m / 106m)   (112Mi / 112Mi)     wrk6 2%
kube-system   kube-dns-7588d5b5f5-8gzbp                 kubedns                  (100m / 0m)     (70Mi / 170Mi)      wrk4 1%
kube-system   kube-dns-7588d5b5f5-n4hqv                 kubedns                  (100m / 0m)     (70Mi / 170Mi)      wrk4 1%
kube-system   kube-state-metrics-679d95df65-bp7cp       addon-resizer            (100m / 100m)   (30Mi / 30Mi)       wrk6 1%
kube-system   kube-dns-autoscaler-5db9bbb766-kkddl      autoscaler               (20m / 0m)      (10Mi / 0Mi)        wrk4 0%
ingress-nginx default-http-backend-797c5bc547-v8dgm     default-http-backend     (10m / 10m)     (20Mi / 20Mi)       wrk4 0%
.
.
.
```
There's not a lot going on in my dev cluster. As you can see, the largest resource requests are by platform services. It might be more interesting to look at only pods in the namespace where application services are running.
```bash
~ » kubestate --namespace default top pods
Namespace Pod                            Container   CPU (Req / Lim) Memory  (Req / Lim) Node Load
default   productpage-v1-54b8b9f55-znw8l istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk6 0%
default   ratings-v1-7bc85949-q6csf      istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk6 0%
default   reviews-v1-fdbf674bb-gqm8f     istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk6 0%
default   reviews-v3-dd846cc78-tpg9v     istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk4 0%
default   httpbin-665f4c5c56-g759l       istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk5 0%
default   reviews-v2-5bdc5877d6-q7xc2    istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk5 0%
default   sleep-5967ffd788-l5czj         istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk5 0%
default   details-v1-6764bbc7f7-698x9    istio-proxy (10m / 0m)      (0Mi / 0Mi)         wrk6 0%
```
It might also be useful to roll up resources by node to see if some nodes are over or under allocated. Remember, these are requested values not actual utilization.
```bash
~ » kubestate top nodes
Node CPU (Req / Lim / Cap)  Memory (Req / Lim / Cap)   Load
wrk6 (1086m / 206m / 4000m) (2190Mi / 142Mi / 15877Mi) 21%
wrk4 (870m / 10m / 4000m)   (250Mi / 360Mi / 15877Mi)  12%
wrk5 (410m / 0m / 4000m)    (0Mi / 0Mi / 15877Mi)      5%
wrk3 (250m / 0m / 4000m)    (0Mi / 0Mi / 15877Mi)      3%
wrk1 (250m / 0m / 4000m)    (0Mi / 0Mi / 15877Mi)      3%
wrk2 (250m / 0m / 4000m)    (0Mi / 0Mi / 15877Mi)      3%
```
To explore further, you can browse through the full list of metrics provided by kube-state-metrics using the kubestate list command.
```bash
~ » kubestate list
GAUGE	kube_configmap_created	Unix creation timestamp
GAUGE	kube_configmap_info	Information about configmap.
.
.
.
GAUGE	kube_pod_container_info	Information about a container in a pod.
GAUGE	kube_pod_container_resource_limits	The number of requested limit resource by a container.
GAUGE	kube_pod_container_resource_limits_cpu_cores	The limit on cpu cores to be used by a container.
GAUGE	kube_pod_container_resource_limits_memory_bytes	The limit on memory to be used by a container in bytes.
GAUGE	kube_pod_container_resource_requests	The number of requested request resource by a container.
GAUGE	kube_pod_container_resource_requests_cpu_cores	The number of requested cpu cores by a container.
GAUGE	kube_pod_container_resource_requests_memory_bytes	The number of requested memory bytes by a container.
.
.
.
```
Then, get the metric values for one using the kubestate get command with the metric filter flag. The jq command is a great way to format the json output to make it easier to read.
```bash
~ » kubestate get -m kube_pod_container_resource_requests_memory_bytes | jq
{
  "name": "kube_pod_container_resource_requests_memory_bytes",
  "help": "The number of requested memory bytes by a container.",
  "type": "GAUGE",
  "metric": [
    {
      "label": [
        {
          "name": "container",
          "value": "addon-resizer"
        },
        {
          "name": "namespace",
          "value": "kube-system"
        },
        {
          "name": "node",
          "value": "wrk6"
        },
        {
          "name": "pod",
          "value": "kube-state-metrics-679d95df65-bp7cp"
        }
      ],
      "gauge": {
        "value": 31457280
      }
    },
    {
      "label": [
        {
          "name": "container",
          "value": "autoscaler"
        },
        {
          "name": "namespace",
          "value": "kube-system"
        },
        {
          "name": "node",
          "value": "wrk4"
        },
        {
          "name": "pod",
          "value": "kube-dns-autoscaler-5db9bbb766-kkddl"
        }
      ],
      "gauge": {
        "value": 10485760
      }
    },
    {
      "label": [
        {
          "name": "container",
          "value": "default-http-backend"
        },
        {
          "name": "namespace",
          "value": "ingress-nginx"
        },
        {
          "name": "node",
          "value": "wrk4"
        },
        {
          "name": "pod",
          "value": "default-http-backend-797c5bc547-v8dgm"
        }
      ],
      "gauge": {
        "value": 20971520
      }
    },
    {
      "label": [
        {
          "name": "container",
          "value": "discovery"
        },
        {
          "name": "namespace",
          "value": "istio-system"
        },
        {
          "name": "node",
          "value": "wrk6"
        },
        {
          "name": "pod",
          "value": "istio-pilot-7d6549448f-6hkvn"
        }
      ],
      "gauge": {
        "value": 2147483648
      }
    },
.
.
.

```
Or, get the output in raw metric exposition format. The Prometheus project has a good description of metric [exposition format](https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md).
```bash
~ » kubestate get -o raw
# HELP kube_configmap_created Unix creation timestamp
# TYPE kube_configmap_created gauge
kube_configmap_created{configmap="canal-config",namespace="kube-system"} 1.536336806e+09
kube_configmap_created{configmap="cluster-state",namespace="kube-system"} 1.536336789e+09
kube_configmap_created{configmap="extension-apiserver-authentication",namespace="kube-system"} 1.536336712e+09
kube_configmap_created{configmap="ingress-controller-leader-nginx",namespace="ingress-nginx"} 1.536337414e+09
kube_configmap_created{configmap="istio",namespace="istio-system"} 1.536620804e+09
.
.
.
```