apt-mark unhold kubeadm kubectl kubelet && \
apt-get update && apt-get install -y --allow-downgrades kubeadm=1.20.15-00 kubectl=1.20.15-00 kubelet=1.20.15-00 && \
apt-mark hold kubeadm

kubeadm upgrade plan

kubeadm upgrade apply v1.20.15