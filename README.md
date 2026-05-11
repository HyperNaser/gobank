# Basic Commands

Create gobank cluster via:
```bash
make cluster
```

Delete gobank cluster via:
```bash
make delete_cluster_gobank
```

Make initial setup via:
```bash
make setup
```

Create postgres database cluster credentials via: 
```bash
make create_creds
```
or manual via (make sure namespace exists first):
```bash
kubectl create secret generic db-creds --namespace gobank --from-literal=username=root --from-literal=password='YOUR_PASSWORD_HERE'
```

Create namespace via:
```bash
make namespace
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

Apply all cluster configs:
```bash
make apply_cluster_configs
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

Generate db docs via:
```bash
tbls doc
```

Generate swagger docs via:
```bash
swag init -g main.go -o docs
```

Make db migration via:
```bash
migrate create -ext sql -dir db/migration -seq <migration_name>
```