Create gobank cluster via:
```bash
make cluster_gobank
```

Create postgres database cluster credentials via: 
```bash
kubectl create secret generic db-creds --from-literal=username=root --from-literal=password='<PASSWORD>'
```

Create postgres database cluster via:
```bash
make cluster_db
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