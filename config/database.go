package config

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/BurntSushi/toml"
	"log"
)

type DBConfig struct {
	DB_Host, DB_User, DB_Password, DB_Name string
	DB_Port                                int64
}

func DBConnection() *sql.DB {

	var dbconfig DBConfig

	_, err := toml.DecodeFile(".env.toml", &dbconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbconfig.DB_Host, dbconfig.DB_Port, dbconfig.DB_User, dbconfig.DB_Password, dbconfig.DB_Name)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err.Error())
	}

	return db
}