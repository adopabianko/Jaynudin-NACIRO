package config

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
)

func DBConnection() *sql.DB {
	viper.SetConfigName(".app-config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	DBhost := viper.Get("db.host")
	DBport := viper.Get("db.port")
	DBname := viper.Get("db.name")
	DBuser := viper.Get("db.user")
	DBpassword := viper.Get("db.password")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBhost, DBport, DBuser, DBpassword, DBname)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err.Error())
	}

	return db
}