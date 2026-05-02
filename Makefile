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
	kubectl apply -f postgres-cluster.yaml

cluster_gobank_api:
	kubectl apply -f gobank-api-deployment.yaml

cluster_gobank:
	k3d cluster create gobank-cluster --api-port 6550 -p "80:80@loadbalancer" -p "443:443@loadbalancer" --agents 2

db_tunnel:
	kubectl port-forward svc/gobank-db-rw 5432:5432

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test mock server migrateup1 migratedown1 cluster_db cluster_start cluster_stop cluster_gobank cluster_gobank_api db_tunnel
