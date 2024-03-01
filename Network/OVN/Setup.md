## Install
- wget https://raw.githubusercontent.com/kubeovn/kube-ovn/release-1.12/dist/images/install.sh
- nano install.sh #Change POD_CIDR,POD_GATEWAY,EXCLUDE_IPS
- bash install.sh

- nano external-gw.yaml
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: ovn-external-gw-config
  namespace: kube-system
data:
  enable-external-gw: "true"
  external-gw-nodes: "kube-ovn-worker"
  external-gw-nic: "eth1"
  external-gw-addr: "172.56.0.1/16"
  nic-ip: "172.56.0.254/16"
  nic-mac: "16:52:f3:13:6a:25"
```
- kubectl ko nbctl show
- kubectl ko vsctl ${gateway node name} show

### Test
- kubectl annotate pod virt-launcher-testvm- ovn.kubernetes.io/eip=172.56.0.221 --overwrite
- kubectl annotate pod virt-launcher-testvm- ovn.kubernetes.io/routed-



### Tshoot
- kubectl ko nbctl lrp-add ovn-cluster ovn-cluster-external 16:52:f3:13:6a:25 192.168.100.10/24
- kubectl ko nbctl lrp-del ovn-cluster-externa
- kubectl run curl --image rancher/curl --command sleep 1d -n another
- kubectl run pipy --image flomesh/pipy:latest -n default
- kubectl ko ofctl ubuntu-kube-1 dump-flows br-int 
