postgres:
	docker run --name gobank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -d postgres:alpine

createdb:
	docker exec -it gobank createdb --username=root --owner=root gobank

dropdb:
	docker exec -it gobank dropdb gobank

migrateup:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -coverprofile=coverage.out ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/HyperNaser/gobank/db/sqlc Store

server:
	go run main.go

cluster_stop:
	k3d cluster stop gobank-cluster

cluster_start:
	k3d cluster start gobank-cluster

cluster_db:
	kubectl apply -f ./k8s/apps/gobank/postgres-cluster.yaml --server-side

cluster_gobank_api:
	kubectl apply -f ./k8s/apps/gobank/gobank-api-deployment.yaml --server-side

apply_cluster_configs: # this is really hacky but we'll roll with it for now. If an issue arises, then rerun the command
	kubectl apply -k ./k8s/base/cnpg-operator --server-side
	sleep 20
	kubectl wait --for condition=established --timeout=60s crd/clusters.postgresql.cnpg.io
	kubectl apply -k ./k8s/apps/gobank --server-side

cluster:
	k3d cluster create --config ./k3d-config.yaml

delete_cluster_gobank:
	k3d cluster delete gobank-cluster

namespace:
	kubectl apply -f k8s/apps/gobank/namespace.yaml

create_creds: namespace
	@read -p "Enter DB Password: " pwd; \
	kubectl create secret generic db-creds \
		--from-literal=username=root \
		--from-literal=password=$$pwd \
		--namespace gobank \
		--dry-run=client -o yaml | kubectl apply -f -

setup: create_creds apply_cluster_configs

db_tunnel:
	kubectl port-forward svc/gobank-db-rw 5432:5432

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test mock server migrateup1 migratedown1 cluster_db cluster_start cluster_stop cluster cluster_gobank_api db_tunnel delete_cluster_gobank apply_cluster_configs namespace create_creds setup
