# Kubernetes Svc Network (calico)

## Cluster IP
- `nginx_svc_ip=$(kubectl get svc -o json | jq -r '.items | .[] | select(.metadata.name|test("nginx.")) | .spec.clusterIP')`
- `ping -c 3 $nginx_svc_ip` #should be timeout
- `curl -sI $nginx_svc_ip | head -n1` #should be 200 ~~daijoubu~~ ok
- `iptables -t nat -nvL PREROUTING`
- `iptables -t nat -nvL KUBE-SERVICES | grep $nginx_svc_ip`
- `iptables -t nat -nvL KUBE-SVC-XXXXX`
- `iptables -t nat -nvL KUBE-SEP-XXXXX` #you should see the pods ip
- `kubectl scale deployment/nginx-deployment --replicas=3`
- `iptables -t nat -nvL KUBE-SVC-XXXXX` #probability traffic should be showing
