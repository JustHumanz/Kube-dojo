# Kubernetes Nginx Ingress

## Ingress

Create Kano-svc
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: kano-index
data:
  index.html: "<html><head><title>Kano</title></head><body>Kano/鹿乃</body></html>"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kano-deployment
spec:
  selector:
    matchLabels:
      app: kano-app
  replicas: 1
  template:
    metadata:
      labels:
        app: kano-app
    spec:
      volumes:
        - name: kano-shared-logs
          emptyDir: {}
        - name: nginx-index-cm
          configMap:
            name: kano-index
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
        volumeMounts:
          - name: kano-shared-logs
            mountPath: /var/log/nginx
          - name: nginx-index-cm
            mountPath: /usr/share/nginx/html/

      - name: nginx-sidecar-container
        image: busybox
        command: ["sh","-c","while true; do cat /var/log/nginx/access.log; sleep 30; done"]
        volumeMounts:
          - name: kano-shared-logs
            mountPath: /var/log/nginx
            readOnly: true
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: kano-svc
  namespace: default
  labels:
    app: kano-app
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: kano-app
  type: ClusterIP
EOF
```


Create Lon-svc
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: lon-index
data:
  index.html: "<html><head><title>Lon</title></head><body>Lon/ろん</body></html>"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lon-deployment
spec:
  selector:
    matchLabels:
      app: lon-app
  replicas: 1
  template:
    metadata:
      labels:
        app: lon-app
    spec:
      volumes:
        - name: lon-shared-logs
          emptyDir: {}
        - name: nginx-index-cm
          configMap:
            name: lon-index
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
        volumeMounts:
          - name: lon-shared-logs
            mountPath: /var/log/nginx
          - name: nginx-index-cm
            mountPath: /usr/share/nginx/html/

      - name: nginx-sidecar-container
        image: busybox
        command: ["sh","-c","while true; do cat /var/log/nginx/access.log; sleep 30; done"]
        volumeMounts:
          - name: lon-shared-logs
            mountPath: /var/log/nginx
            readOnly: true
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: lon-deployment
  namespace: default
  labels:
    app: lon-app
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: lon-app
  type: ClusterIP
EOF
```

Create ingress

```bash
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mklntic-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  defaultBackend:
    service:
      name: kano-svc
      port:
        number: 80  
  rules:
  - host: "mklntic.moe"
    http:
      paths:
      - path: /kano
        pathType: Prefix
        backend:
          service:
            name: kano-svc
            port:
              number: 80  
      - path: /lon
        pathType: Prefix
        backend:
          service:
            name: lon-svc
            port:
              number: 80
EOF
```

- `ingress_ip=$(kubectl get svc --namespace=ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}')`
- `curl -H "Host: mklntic.moe" $ingress_ip` #the output should be **Kano/鹿乃**
- `curl -H "Host: mklntic.moe" $ingress_ip/lon` #the output should be **Lon/ろん**
- `curl -H "Host: mklntic.moe" $ingress_ip/kano` #the output should be **Kano/鹿乃**