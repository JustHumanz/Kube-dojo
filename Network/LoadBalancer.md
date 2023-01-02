# Kubernetes Svc Network (calico)

## LoadBalancer
- `nginx_svc_lb=$(kubectl get svc -ojson | jq -r '.items | .[] | select(.metadata.name|test("nginx.")) | .status.loadBalancer.ingress[0].ip')`
- `ping -c 3 $nginx_svc_lb`
- `curl $nginx_svc_lb` #curl from inside kube cluster
- `curl $nginx_svc_lb` #curl from outside kube cluster
- `arp -an` #nginx_svc_lb should have same arp
- `iptables -t nat -nvL | grep $nginx_svc_lb`
- `iptables -t nat -nvL KUBE-FW-XXXXX` #KUBE-FW should have KUBE-SVC like in Cluster_IP.md

##### Test externalTrafficPolicy local
- `kubectl edit svc/nginx-deployment` #Add `externalTrafficPolicy: Local`
- `iptables -t nat -L | grep $nginx_svc_lb` #output should be empty
- `iptables -t nat -L | grep $nginx_svc_lb` #exec on ubuntu-nested-3 and the iptables should not empty