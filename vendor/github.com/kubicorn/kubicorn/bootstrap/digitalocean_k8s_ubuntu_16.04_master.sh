# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use.
KUBERNETES_VERSION="1.11.1"
KUBERNETES_CNI="0.6.0"
DOCKER_VERSION="17.03"

# Obtain Droplet IP addresses.
HOSTNAME=$(curl -s http://169.254.169.254/metadata/v1/hostname)
PRIVATEIP=$(curl -s http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address)
PUBLICIP=$(curl -s http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address)

# Add Kubernetes repository.
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
touch /etc/apt/sources.list.d/kubernetes.list
sh -c 'echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

# Add Docker repository
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sh -c 'echo "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list'

# Update apt cache
apt-get update -y

# Get docker version
pkg_pattern="$(echo "$DOCKER_VERSION" | sed "s/-ce-/~ce~/g" | sed "s/-/.*/g").*-0~ubuntu"
search_command="apt-cache madison 'docker-ce' | grep '$pkg_pattern' | head -1 | cut -d' ' -f 4"
pkg_version="$(sh -c "$search_command")"

# Install packages.
apt-get install -y \
    socat \
    ebtables \
    docker-ce="${pkg_version}" \
    apt-transport-https \
    kubelet=${KUBERNETES_VERSION}-00 \
    kubeadm=${KUBERNETES_VERSION}-00 \
    kubernetes-cni=${KUBERNETES_CNI}-00 \
    cloud-utils \
    jq

# Enable and start Docker.
systemctl enable docker
systemctl start docker

# Parse kubicorn configuration file.
CLUSTER_NAME=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.metadata.name')
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
PORT=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDPORT | tonumber')

# Create kubeadm configuration file.
touch /etc/kubicorn/kubeadm-config.yaml
cat << EOF  > "/etc/kubicorn/kubeadm-config.yaml"
apiVersion: kubeadm.k8s.io/v1alpha2
kind: MasterConfiguration
bootstrapTokens:
- token: ${TOKEN}
kubernetesVersion: ${KUBERNETES_VERSION}
nodeName: ${HOSTNAME}
clusterName: ${CLUSTER_NAME}
api:
  advertiseAddress: ${PUBLICIP}
  bindPort: ${PORT}
apiServerCertSANs:
- ${PRIVATEIP}
- ${PUBLICIP}
- ${HOSTNAME}
authorizationModes:
- Node
- RBAC
networking:
  podSubnet: "10.244.0.0/16"
EOF

# Initialize cluster.
kubeadm reset --force
kubeadm init --config /etc/kubicorn/kubeadm-config.yaml

# Flannel CNI plugin
sysctl net.bridge.bridge-nf-call-iptables=1
curl -SL "https://raw.githubusercontent.com/coreos/flannel/v0.10.0/Documentation/kube-flannel.yml" \
 | kubectl apply --kubeconfig /etc/kubernetes/admin.conf -f -

# DigitalOcean Cloud-Manager
curl -SL "https://raw.githubusercontent.com/digitalocean/digitalocean-cloud-controller-manager/master/releases/v0.1.7.yml" \
 | kubectl apply --kubeconfig /etc/kubernetes/admin.conf -f -
curl -SL "https://raw.githubusercontent.com/digitalocean/csi-digitalocean/master/deploy/kubernetes/releases/csi-digitalocean-v0.2.0.yaml" \
| kubectl apply --kubeconfig /etc/kubernetes/admin.conf -f -

mkdir -p /root/.kube
cp /etc/kubernetes/admin.conf /root/.kube/config
chown -R root:root /root/.kube
