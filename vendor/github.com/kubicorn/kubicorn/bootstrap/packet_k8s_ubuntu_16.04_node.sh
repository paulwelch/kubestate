# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use.
KUBERNETES_VERSION="1.10.4"
KUBERNETES_CNI="0.6.0"

# Controls delay before attempting to join the master
MAX_ATTEMPTS=50
REATTEMPT_INTERVAL_SECONDS=30

# order is important:
# 1. update
# 2. install apt-transport-https
# 3. add kubernetes repos to list
# 4. update again
# 5. install
apt-get update -y
apt-get install -y apt-transport-https

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
touch /etc/apt/sources.list.d/kubernetes.list
sh -c 'echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

apt-get update -y

apt-get install -y \
    socat \
    ebtables \
    docker.io \
    apt-transport-https \
    kubelet=${KUBERNETES_VERSION}-00 \
    kubeadm=${KUBERNETES_VERSION}-00 \
    kubernetes-cni=${KUBERNETES_CNI}-00 \
    jq

systemctl enable docker
systemctl start docker

# must disable swap for kubelet to work
swapoff -a


TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
MASTER=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDMASTER')

systemctl daemon-reload
systemctl restart kubelet.service

# Reset before joining
kubeadm reset

# Delay kubeadm join until master is ready
attempts=0
response=000
while [ "${response}" -ne "200" ] && [ $(( attempts++ )) -lt $MAX_ATTEMPTS ]; do
  echo "Waiting for master to be ready(${MASTER})..."
  sleep $REATTEMPT_INTERVAL_SECONDS
  response=$(curl --write-out "%{http_code}" --output /dev/null --silent --connect-timeout 10 -k "https://${MASTER}/healthz" || true)
done

# Join the cluster
if [ "${response}" -ne "200" ]; then
  echo "Maximum attempts reached, giving up"
  exit 1
else
  echo "Master seems to be up and running. Joining the node to the cluster..."
  kubeadm join --node-name "${HOSTNAME}" --token "${TOKEN}" "${MASTER}" --discovery-token-unsafe-skip-ca-verification
fi