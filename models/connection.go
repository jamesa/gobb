package models

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	"github.com/stevenleeg/gobb/config"
)

var dbMap *gorp.DbMap

func GetDbSession() *gorp.DbMap {
	if dbMap != nil {
		return dbMap
	}

	db_username, _ := config.Config.GetString("database", "username")
	db_password, _ := config.Config.GetString("database", "password")
	db_database, _ := config.Config.GetString("database", "database")
	db_hostname, _ := config.Config.GetString("database", "hostname")
	db_port, _ := config.Config.GetString("database", "port")

	db_env_hostname, _ := config.Config.GetString("database", "env_hostname")
	db_env_port, _ := config.Config.GetString("database", "env_port")

	// Allow database information to come from environment variables
	if db_env_hostname != "" {
		db_hostname = os.Getenv(db_env_hostname)
	}
	if db_env_port != "" {
		db_port = os.Getenv(db_env_port)
	}

	if db_port == "" {
		db_port = "5432"
	}

	db, err := sql.Open("postgres",
		"user="+db_username+
			" password="+db_password+
			" dbname="+db_database+
			" host="+db_hostname+
			" port="+db_port+
			" sslmode=disable")

	if err != nil {
		fmt.Printf("Cannot open database! Error: %s\n", err.Error())
		return nil
	}

	dbMap = &gorp.DbMap{
		Db:      db,
		Dialect: gorp.PostgresDialect{},
	}

	// TODO: Do we need this every time?
	dbMap.AddTableWithName(User{}, "users").SetKeys(true, "ID")
	dbMap.AddTableWithName(Board{}, "boards").SetKeys(true, "ID")
	dbMap.AddTableWithName(Post{}, "posts").SetKeys(true, "ID")
	dbMap.AddTableWithName(View{}, "views").SetKeys(false, "ID")
	dbMap.AddTableWithName(Setting{}, "settings").SetKeys(true, "Key")

	return dbMap
}
