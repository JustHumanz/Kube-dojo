```bash
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: kano-app
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: lon-app
    ports:
      - port: 80
EOF
```