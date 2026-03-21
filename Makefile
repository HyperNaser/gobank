postgres:
	docker run --name gobank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -d postgres:alpine

createdb:
	docker exec -it gobank createdb --username=root --owner=root gobank

dropdb:
	docker exec -it gobank dropdb gobank

.PHONY: postgres createdb dropdb
