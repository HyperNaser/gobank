postgres:
	docker run --name gobank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -d postgres:alpine

createdb:
	docker exec -it gobank createdb --username=root --owner=root gobank

dropdb:
	docker exec -it gobank dropdb gobank

migrateup:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration/ -database "postgresql://root:root@localhost:5432/gobank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown sqlc
