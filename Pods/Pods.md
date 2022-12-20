# Pods

Deep down whats is pods and how is running in kubernetes

```bash
cat <<EOF | kubectl apply -f -
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
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        volumeMounts:
          - name: shared-logs
            mountPath: /var/log/nginx
        - containerPort: 80

      - name: nginx-sidecar-container
        image: busybox
        command: ["sh","-c","while true; do cat /var/log/nginx/access.log; sleep 30; done"]
        volumeMounts:
          - name: shared-logs
            mountPath: /var/log/nginx

      nodeSelector:
        kubernetes.io/hostname: "ubuntu-nested-3"
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
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
  type: NodePort
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
