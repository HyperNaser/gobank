Create gobank cluster via:
# Basic Commands

```bash
make cluster_gobank
```

Create postgres database cluster credentials via: 
```bash
kubectl create secret generic db-creds --from-literal=username=root --from-literal=password='YOUR_PASSWORD_HERE'
```

Delete postgres database cluster credentials via:
```bash
kubectl delete secret db-creds
```

Check pods via:
```bash
kubectl get pods
```

Delete a deployment via:
```bash
kubectl delete deployment DEPLOYMENT_NAME
```

Apply a deployment yaml file via:
```bash
kubectl apply -f PATH_TO_YAML
```

Create postgres database cluster via:
```bash
make cluster_db
```

Create gobank-api cluster via:
```bash
make cluster_gobank_api
```

Start cluster via:
```bash
make cluster_start
```

Stop cluster via:
```bash
make cluster_stop
```

Expose a port to database via:
```bash
kubectl port-forward svc/gobank-db-rw 5432:5432
```