Create gobank cluster via:
```bash
make cluster_gobank
```

Create postgres database cluster credentials via: 
```bash
make create_secret PWD='the_password'
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