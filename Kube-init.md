# Topology
```
                                   +---------+
                                   |  My PC  |
        +--------------------------+         +--------------------------+
        |                          +---------+                          |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        v                               v                               v
+-----------------+            +-----------------+           +-----------------+
| Ubuntu nested 1 |            | Ubuntu nested 2 |           | Ubuntu nested 3 |
|   200.0.0.10    |            |   200.0.0.20    |           |   200.0.0.30    |
|    Master 1     |            |    Worker 1     |           |    Worker 2     |
|                 |            |                 |           |                 |
+-----------------+            +-----------------+           +-----------------+
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        |                               |                               |
        +---------------------------------------------------------------+
                         Kube cluster network (200.0.0.0/24)
```

# Setup Containerd
- `wget https://github.com/containerd/containerd/releases/download/v1.6.2/containerd-1.6.2-linux-amd64.tar.gz`
- `sudo tar Czxvf /usr/local containerd-1.6.2-linux-amd64.tar.gz`

- `wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service`
- `sudo mv containerd.service /usr/lib/systemd/system/`

- `sudo systemctl daemon-reload`
- `sudo systemctl enable --now containerd`
- `sudo systemctl status containerd`

- `wget https://github.com/opencontainers/runc/releases/download/v1.1.1/runc.amd64`
- `sudo install -m 755 runc.amd64 /usr/local/sbin/runc`

- `sudo mkdir -p /etc/containerd/`
- `containerd config default | sudo tee /etc/containerd/config.toml`

- `sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml`

- `sudo systemctl restart containerd`

- `cat <<EOF | sudo tee /etc/modules-load.d/containerd.conf \noverlay \nbr_netfilter \nEOF`

- `sudo modprobe overlay `
- `sudo modprobe br_netfilter`

- `cat <<EOF | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf net.bridge.bridge-nf-call-iptables = 1 \nnet.ipv4.ip_forward = 1 \nnet.bridge.bridge-nf-call-ip6tables = 1 \nEOF`

- `sudo sysctl --system`

# Setup Kube
- `sudo apt update && sudo apt-get install -y apt-transport-https curl`
- `curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -`

- `cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list\ndeb https://apt.kubernetes.io/ kubernetes-xenial main\nEOF`

- `sudo apt update`

- `sudo apt-get install -y kubelet=1.17.1-00 kubeadm=1.17.1-00 kubectl=1.17.1-00`
- `sudo apt-mark hold kubelet kubeadm kubectl`

- `nano cluster.yaml`
```yaml
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta2
kubernetesVersion: v1.17.1
controlPlaneEndpoint: "200.0.0.10:6443"
networking:
  podSubnet: "100.100.0.0/16"
---
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
cgroupDriver: systemd
```

- `kubeadm init --config cluster.yaml --cri-socket /var/run/containerd/containerd.sock`

- `kubeadm join 200.0.0.10:6443 --token 7eq77f.14w92ywfymbdvgw5 --discovery-token-ca-cert-hash sha256:e7e11ad25b273e6aa8296cdb68ba32208feec9a396b611f614fa7a6c19bdc570 --cri-socket /var/run/containerd/containerd.sock`

# Setup Calico cni
- `wget https://docs.projectcalico.org/archive/v3.12/manifests/calico.yaml`
- `nano calico.yaml`
- `kubectl apply -f calico.yaml`

# Setup MetalLB
- `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11/manifests/metallb.yaml`
- `nano metallb-system-config.yaml`
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 200.0.0.100-200.0.0.120
```
- `kubectl apply -f metallb-system-config.yaml`