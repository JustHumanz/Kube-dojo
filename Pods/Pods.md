# Pods

Deep down whats is pods and how is running in kubernetes

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-index
data:
  index.html: "<html><head><title>Kano</title></head><body>Kano/鹿乃</body></html>"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 1
  template:
    metadata:
      labels:
        app: nginx
    spec:
      volumes:
        - name: shared-logs
          emptyDir: {}
        - name: nginx-index-cm
          configMap:
            name: nginx-index
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
        volumeMounts:
          - name: shared-logs
            mountPath: /var/log/nginx
          - name: nginx-index-cm
            mountPath: /usr/share/nginx/html/
        resources:
          requests:
            memory: 50Mi
          limits:
            memory: 100Mi

      - name: nginx-sidecar-container
        image: busybox
        command: ["sh","-c","while true; do cat /var/log/nginx/access.log; sleep 30; done"]
        volumeMounts:
          - name: shared-logs
            mountPath: /var/log/nginx
            readOnly: true

      nodeSelector:
        kubernetes.io/hostname: "ubuntu-nested-3"
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-deployment
  namespace: default
  labels:
    app: nginx
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
EOF
```
### Pause

- `kubectl get pods -l app=nginx -o json | jq '.items | .[0] | .status | .containerStatuses | .[0] | .containerID' | sed 's/containerd:\/\///'` #save the id
- `ctr -n k8s.io c info <Container ID> | grep namespaces -A 20` #save the pid
- `ctr -n k8s.io c info $(ctr -n k8s.io c info <Container ID> | jq -r '.Spec | .annotations | ."io.kubernetes.cri.sandbox-id"') | grep cni` #save the cni 
- `ip netns pids <cni ns> | head -n1` #compair the pid,the pid should same 
- `pstree -aps <pid>` #the pid should haved by pause

in summary the pause image was who create the network,uts,ipc namespaces and the nginx container attach the namespaces, so the pods will only request ip address one even if the nginx container was crash or die

### Multi container in single pods
- `kubectl get pods -l app=nginx -o json | jq '.items | .[0] | .status | .containerStatuses | .[] | .containerID' | sed 's/containerd:\/\///'`
- `ctr -n k8s.io c info <Container ID 0> | grep namespaces -A 20`
- `ctr -n k8s.io c info <Container ID 1> | grep namespaces -A 20` 
the namespace pid should be same
- `ctr -n k8s.io c info <Container ID 0> | grep shared`
- `ctr -n k8s.io c info <Container ID 1> | grep shared`
the mounting dir should be same


### Resources
- `kubectl get pods -l app=nginx -o json | jq -r '.items | .[0] | .status.qosClass, .metadata.uid'` #save the qosClass & uid
- `kubectl get pods -l app=nginx -o json | jq '.items | .[0] | .status | .containerStatuses | .[] | .containerID' | sed 's/containerd:\/\///'` #save the containerID
- `cd /sys/fs/cgroup/memory/kubepods/echo <qosClass> | | tr '[:upper:]' '[:lower:]'/pod<uid>`
- `cat <container ID 0>/memory.limit_in_bytes`
- `cat <container ID 1>/memory.limit_in_bytes`
One of them should have 104857600
- `python -c 'print(104857600/1024/1024)'` #the output should be 100