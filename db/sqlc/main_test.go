package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/HyperNaser/gobank/util"
	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
	testDB      *sql.DB
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Could not load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
