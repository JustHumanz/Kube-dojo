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
- `mkdir -m 0755 -p /etc/apt/keyrings`
- `curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg`
- `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`
- `apt update;apt-get install containerd.io -y`
- `containerd config default | sudo tee /etc/containerd/config.toml`
- `sudo echo -e "noverlay\nbr_netfilter" > /etc/modules-load.d/containerd.conf`

- `sudo modprobe overlay `
- `sudo modprobe br_netfilter`
- `sudo echo -e "net.bridge.bridge-nf-call-iptables = 1\nnet.ipv4.ip_forward = 1\nnet.bridge.bridge-nf-call-ip6tables = 1" > /etc/sysctl.d/99-kubernetes-cri.conf`

- `nano /etc/containerd/config.toml` #set SystemdCgroup to false
- `sudo sysctl --system`
- `systemctl restart containerd`

# Setup Kube
- `sudo apt update && sudo apt-get install -y apt-transport-https curl`
- `curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -`
- `sudo echo -e "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list`

- `sudo apt update`
- `sudo apt-get install -y kubelet=1.20.15-00 kubeadm=1.20.15-00 kubectl=1.20.15-00`
- `sudo apt-mark hold kubelet kubeadm kubectl`

- `nano cluster.yaml`
```yaml
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta2
kubernetesVersion: v1.20.15
controlPlaneEndpoint: "200.0.0.10:6443"
networking:
  podSubnet: "100.100.0.0/16"
---
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
```

- `kubeadm init --config cluster.yaml --cri-socket /var/run/containerd/containerd.sock`

- `kubeadm join 200.0.0.10:6443 --token 7eq77f.14w92ywfymbdvgw5 --discovery-token-ca-cert-hash sha256:e7e11ad25b273e6aa8296cdb68ba32208feec9a396b611f614fa7a6c19bdc570 --cri-socket /var/run/containerd/containerd.sock`

# Setup Calico cni
- `wget https://docs.projectcalico.org/archive/v3.12/manifests/calico.yaml`
- `nano calico.yaml`
- `kubectl apply -f calico.yaml`

# Setup MetalLB
- `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml`
- `nano metallb-system-config.yaml`
```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: default
  namespace: metallb-system
spec:
  addresses:
  - 200.0.0.100-200.0.0.120
  autoAssign: true
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: default
  namespace: metallb-system
spec:
  ipAddressPools:
  - default
```
- `kubectl apply -f metallb-system-config.yaml`

# Ingress
- `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.3.1/deploy/static/provider/cloud/deploy.yaml`


# Ref 
- https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/
- https://kubernetes.github.io/ingress-nginx/deploy/
- https://medium.com/google-cloud/understanding-kubernetes-networking-ingress-1bc341c84078
- https://medium.com/google-cloud/understanding-kubernetes-networking-services-f0cb48e4cc82
- https://medium.com/@betz.mark/understanding-kubernetes-networking-pods-7117dd28727
- https://gist.github.com/mcastelino/c38e71eb0809d1427a6650d843c42ac2#targets
- https://www.tkng.io/services/clusterip/dataplane/iptables/
- https://man7.org/linux/man-pages/man8/ip-netns.8.html
- https://unix.stackexchange.com/questions/213054/how-to-list-processes-belonging-to-a-network-namespace
- https://blog.quarkslab.com/digging-into-linux-namespaces-part-2.html
- https://blog.quarkslab.com/digging-into-runtimes-runc.html
- https://man7.org/linux/man-pages/man2/clone.2.html
- https://www.xiemx.com/2019/09/16/k8s-ingress-nginx/index.html
- https://www.asykim.com/blog/deep-dive-into-kubernetes-external-traffic-policies
- https://andys.org.uk/bits/2010/01/27/iptables-fun-with-mark/comment-page-1/
- https://lwn.net/Articles/532593/
- https://github.com/teddyking/ns-process
- https://www.altoros.com/blog/kubernetes-networking-writing-your-own-simple-cni-plug-in-with-bash/
- https://github.com/s-matyukevich/bash-cni-plugin/blob/master/01_gcp/bash-cni
- https://dustinspecker.com/posts/kubernetes-networking-from-scratch-bgp-bird-advertise-pod-routes/