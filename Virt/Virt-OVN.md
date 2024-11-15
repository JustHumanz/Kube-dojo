### Setup
first setup [kubeovn](../Network/OVN/Setup.md)

## Create ext-port
- openstack port create --network public_1 kubevirt-ovn-ext
- openstack port show kubevirt-ovn-ext #save ip addr&mac
- openstack port set --fixed-ip subnet=public_1,ip-address=X.X.X.X kubevirt-ovn-ext

## Add ext gatewat
- nano external-gw.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ovn-external-gw-config
  namespace: kube-system
data:
  enable-external-gw: "true"
  external-gw-nodes: "XXXXX"
  external-gw-nic: "ensX"
  external-gw-addr: "X.X.X.254/24"
  nic-ip: "X.X.X.X/24" #ip from openstack
  nic-mac: "X:X:X:X:X:X" #mac from openstack
```
