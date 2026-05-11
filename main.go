// @title GoBank API
// @version 1.0
// @description GoBank is a simple banking API built with Go, Gin, and SQLC.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@localhost
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"database/sql"
	"log"

	"github.com/HyperNaser/gobank/api"
	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/docs"
	"github.com/HyperNaser/gobank/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if config.ServerAddress == "0.0.0.0:8080" {
		docs.SwaggerInfo.Host = "localhost:8080"
	} else {
		docs.SwaggerInfo.Host = config.ServerAddress
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
