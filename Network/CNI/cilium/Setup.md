## Remove calico

```bash
kubectl delete crd $(kubectl get crd | grep calico | awk '{print $1}')
kubectl delete -n kube-system deployment calico-kube-controllers
kubectl delete -n kube-system daemonset calico-node
kubectl -n kube-system delete pod/calico-kube-controllers-xxxxxx
```

## Install cilium

```bash
helm install cilium cilium/cilium --version 1.16.2 --namespace kube-system -f values.yaml
cilium status --wait
reboot #all worker nodes
```

```bash
iptables-save | grep SVC #empty
```
## Create dummy svc
```bash
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Service
metadata:
  name: humanz-app
  labels:
    app: humanz-app
spec:
  type: ClusterIP
  ports:
    - port: 80
  selector:
    app: humanz-nginx

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: humanz-nginx
spec:
  selector:
    matchLabels:
      app: humanz-nginx
  replicas: 6
  template:
    metadata:
      labels:
        app: humanz-nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
EOF        
```

## ebpf
[Read my post about ebpf](https://humanz.moe/posts/into-eBPF-Xdp&Tc/)



## cgroup/connect4

Create pods

```bash
WORKER_X=$(kubectl get nodes | grep worker | head -n1 | awk '{print $1}')
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: netshoot-1
spec:
  nodeName: $WORKER_X
  containers:
  - name: netshoot-1
    image: nicolaka/netshoot
    imagePullPolicy: IfNotPresent
    command: ["/bin/bash","-c", "while true; do ping localhost; sleep 60;done"]
EOF
```

Split terminal into 2

#### terminal 1
```bash
kubectl get svc -l app=humanz-app #save the cluster ip
kubectl exec -it pods/netshoot-1 -- bash
ip link show eth0 | grep -oE 'if[0-9]{0,}' | cut -c 3- #save the interface_id and continue exec on terminal 2
curl http://$humanz-app-cluster-ip #continue exec on terminal 2
nc -v -w10 $humanz-app-cluster-ip 80& 
ss -tupn #the "Peer Address:Port" should be one of humanz-nginx pods ip
```
#### terminal 2
```bash
kubectl -n kube-system exec -it pods/$(kubectl get pods -A -l app.kubernetes.io/name=cilium-agent -o=custom-columns=NODE:.spec.nodeName,NAME:.metadata.name | grep $WORKER_X | awk '{print $2}') -- bash
apt update;apt install tcpdump -y
ip link | grep $interface_id: #tcpdump the interface name
tcpdump -ni $output_from_ip_link -c 6 #continue exec on terminal 1
.....
.......
........
......... #the output should be one of humanz-nginx pods ip
```

### Trace the ebpf

get svc ip address and port
```bash
╰─(ﾉ˚Д˚)ﾉ kubectl get svc -l app=humanz-app
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
humanz-app   ClusterIP   10.254.224.248   <none>        80/TCP    41m
```

ssh/exec bash into cilium agent/worker then jump into `/sys/fs/bpf/tc/globals/`
```bash
kubectl -n kube-system exec -it ds/cilium -- bash
cd /sys/fs/bpf/tc/globals/
```

let's see the files in this dir
```bash
root@kube-nkdqq-default-worker-vrbms-x4t7h-f7vgf:/sys/fs/bpf/tc/globals# ls -1
cilium_auth_map
cilium_call_policy
cilium_calls_00011
cilium_calls_00255
cilium_calls_00483
cilium_calls_01040
cilium_calls_01093
cilium_calls_hostns_02109
cilium_calls_netdev_00002
cilium_calls_netdev_00003
cilium_calls_netdev_00010
cilium_calls_overlay_2
cilium_ct4_global
cilium_ct_any4_global
cilium_egresscall_policy
cilium_events
cilium_ipcache
cilium_ipv4_frag_datagrams
cilium_l2_responder_v4
cilium_lb4_affinity
cilium_lb4_backends_v3
cilium_lb4_reverse_nat
cilium_lb4_reverse_sk
cilium_lb4_services_v2
cilium_lb4_source_range
cilium_lb_affinity_match
cilium_lxc
cilium_metrics
cilium_node_map
cilium_node_map_v2
cilium_nodeport_neigh4
cilium_policy_00011
cilium_policy_00255
cilium_policy_00483
cilium_policy_01040
cilium_policy_01093
cilium_policy_02109
cilium_ratelimit
cilium_runtime_config
cilium_signals
cilium_skip_lb4
cilium_snat_v4_external
cilium_tunnel_map
```

theres many file with name **cilium_XXXX**, they are ebpf map. for now let's focust on `cilium_lb4_services_v2` since we want to trace lb.

```bash
root@kube-44ae8-default-worker-klxsw-t9swq-9rbrw:/sys/fs/bpf/tc/globals# bpftool map dump pinned cilium_lb4_services_v2 | head
key: 0a fe 00 01 01 bb 00 00  00 00 00 00  value: 00 00 00 00 03 00 00 01  00 00 00 00
key: 0a fe e0 f8 00 50 05 00  00 00 00 00  value: 26 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe 00 0a 23 c1 00 00  00 00 00 00  value: 00 00 00 00 02 00 00 05  00 00 00 00
key: 0a fe e0 f8 00 50 02 00  00 00 00 00  value: 23 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe 00 01 01 bb 02 00  00 00 00 00  value: 03 00 00 00 00 00 00 01  00 00 00 00
key: 0a fe e0 f8 00 50 06 00  00 00 00 00  value: 27 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe 00 0a 23 c1 01 00  00 00 00 00  value: 12 00 00 00 00 00 00 05  00 00 00 00
key: 0a fe 98 da 20 fb 01 00  00 00 00 00  value: 0c 00 00 00 00 00 00 03  00 00 00 00
key: 0a fe 98 da 20 fb 02 00  00 00 00 00  value: 0d 00 00 00 00 00 00 03  00 00 00 00
key: 0a fe a3 ab 01 bb 01 00  00 00 00 00  value: 04 00 00 00 00 00 00 02  00 00 00 00
```
there are two element, key and value or key value object and they are in hexadecimal format. since they are hexadecimal format we need to change our ip into hexadecimal

```python
>>> for i in [10,254,224,248]:
...     print(f'{i:x}',end=" ")
...
... 0a fe e0 f8
```

okey now already convert ip address into hex, let's find it with grep.

```bash
root@kube-44ae8-default-worker-klxsw-t9swq-9rbrw:/sys/fs/bpf/tc/globals# bpftool map dump pinned cilium_lb4_services_v2 | grep "0a fe e0 f8"
key: 0a fe e0 f8 00 50 05 00  00 00 00 00  value: 26 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 02 00  00 00 00 00  value: 23 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 06 00  00 00 00 00  value: 27 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 04 00  00 00 00 00  value: 25 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 01 00  00 00 00 00  value: 22 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 03 00  00 00 00 00  value: 24 00 00 00 00 00 00 08  00 00 00 00
key: 0a fe e0 f8 00 50 00 00  00 00 00 00  value: 00 00 00 00 06 00 00 08  00 00 00 00
```

>>How to read this?

according the struct format, the first 32 byte is for ip address and then 16 bytes for port. we can see the detail of struct in [here](https://github.com/cilium/cilium/blob/8c315f02f3be6278a77af279fd135ec9b26646c2/bpf/lib/common.h#L1061)

[Key struct](https://github.com/cilium/cilium/blob/8c315f02f3be6278a77af279fd135ec9b26646c2/bpf/lib/common.h#L1061)
```c
struct lb4_key {
	__be32 address;		/* Service virtual IPv4 address */
	__be16 dport;		/* L4 port filter, if unset, all ports apply */
	__u16 backend_slot;	/* Backend iterator, 0 indicates the svc frontend */
	__u8 proto;		/* L4 protocol, 0 indicates any protocol */
	__u8 scope;		/* LB_LOOKUP_SCOPE_* for externalTrafficPolicy=Local */
	__u8 pad[2];
};
```

[Value struct](https://github.com/cilium/cilium/blob/8c315f02f3be6278a77af279fd135ec9b26646c2/bpf/lib/common.h#L1070)
```c
struct lb4_service {
	union {
		__u32 backend_id;	/* Backend ID in lb4_backends */
		__u32 affinity_timeout;	/* In seconds, only for svc frontend */
		__u32 l7_lb_proxy_port;	/* In host byte order, only when flags2 && SVC_FLAG_L7LOADBALANCER */
	};
	/* For the service frontend, count denotes number of service backend
	 * slots (otherwise zero).
	 */
	__u16 count;
	__u16 rev_nat_index;	/* Reverse NAT ID in lb4_reverse_nat */
	__u8 flags;
	__u8 flags2;
	/* For the service frontend, qcount denotes number of service backend
	 * slots under quarantine (otherwise zero).
	 */
	__u16 qcount;
};
```

or we can see it like this 

```                                                                              
           ┌───────────────────────────► Service IPv4 Address                               
           │                                                                                
           │                                                                                
           │         ┌─────────────────► Service Port                                       
           │         │                                                                      
           │         │                                                                      
           │         │      ┌──────────► Backend Slot                                       
           │         │      │                                                               
           │         │      │                                                               
     ┌─────┴─────┐┌──┴──┐┌──┴──┐                   ┌───────────┐┌─────┐                   
 key:│0a fe e0 f8││00 50││05 00│00 00 00 00  value:│26 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││02 00│00 00 00 00  value:│23 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││06 00│00 00 00 00  value:│27 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││04 00│00 00 00 00  value:│25 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││01 00│00 00 00 00  value:│22 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││03 00│00 00 00 00  value:│24 00 00 00││00 00│ 00 08  00 00 00 00
 key:│0a fe e0 f8││00 50││00 00│00 00 00 00  value:│00 00 00 00││06 00│ 00 08  00 00 00 00
     └───────────┘└─────┘└─────┘                   └─────┬─────┘└──┬──┘                   
                                                         │         └─────────────►  Count of backend endpoint
                                                         │                                  
                                                         │                                  
                                                         │                                  
                                                         └─────────────────────────►  Backend ID      
```
Ok, let me explain it.

first cilium will take the dst ip addr,port, and protocol

[the code](https://github.com/cilium/cilium/blob/c9e65e6debe4266f15473eb016b15141a78bb50d/bpf/bpf_sock.c#L253C1-L259C20)
```c
	struct lb4_key key = {
		.address	= dst_ip,
		.dport		= dst_port,
#if defined(ENABLE_SERVICE_PROTOCOL_DIFFERENTIATION)
		.proto		= protocol,
#endif
	}, orig_key = key;

....
......
.......
........
.........
	svc = lb4_lookup_service(&key, true);
```

Since cilium only have ip addr,port,and protocol that mean cilium only can resolve one key and that key don't have backend slot or backend id in column map, but in that value the map have `Count of backend endpoint` column. and from that cilium doing the modulus math operator to find/random backend 

[the code](https://github.com/cilium/cilium/blob/c9e65e6debe4266f15473eb016b15141a78bb50d/bpf/bpf_sock.c#L366)
```c
		key.backend_slot = (sock_select_slot(ctx_full) % svc->count) + 1;
		backend_slot = __lb4_lookup_backend_slot(&key);
		if (!backend_slot) {
			update_metrics(0, METRIC_EGRESS, REASON_LB_NO_BACKEND_SLOT);
			return -EHOSTUNREACH;
		}

		backend_id = backend_slot->backend_id;
		backend = __lb4_lookup_backend(backend_id);
```
in example the `Count of backend endpoint` colum was 6 so let we modulus it by random number.

```python
>>> import random
>>> random.randint(0, 4294967295)%6+1
5
```
Now the result of modulus operator will fill the backend slot column and after that cilium will resolve the backend slot.