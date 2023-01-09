# Kubernetes Pods Network (Calico & IP-in-IP)

Keep in mind,my pods ip subnet was `100.100.0.0/16`.

## Pods IP
- `kubectl get pods -o wide -l app=nginx` #make sure the pods was located on ubuntu-nested-3
- `ping -c 3 <pods ip>`
- `tracepath <pods ip>` #the first ip should come from ubuntu-nested-3 
- `ip route | grep <ubuntu-nested-3 ip>`
ssh into ubuntu-nested-3
- `ip route | grep <pods ip>` #the routing table should routed into dev calXXXX
- `ctr -n k8s.io c info $(ctr -n k8s.io c info <Container ID> | jq -r '.Spec | .annotations | ."io.kubernetes.cri.sandbox-id"') | grep cni`
- `ip netns exec <cni id> ethtool -S eth0`
- `ip link | grep <peer_ifindex id>` # the device should same like ip route

# Kubernetes Pods Network (Calico & VXLAN)

## Pods IP
- `kubectl edit ds/calico-node -n kube-system`
- Change `CALICO_IPV4POOL_IPIP` with `CALICO_IPV4POOL_VXLAN`
- Comment `# - -bird-live` under livenessProbe&readinessProbe
- `kubectl delete pods -o wide -l app=nginx` # Recreate the pod
- `kubectl get pods -o wide -l app=nginx`
- `ping -c 3 $POD_IP`
- `ip route | grep $POD_IP`
ssh to ubuntu-nested-3
- `ip route | grep $POD_IP`