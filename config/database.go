package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

const (
	host = "localhost"
	port = 5432
)

func ConnectToDB() *sql.DB {
	DBpassword, DBuser, DBname := os.Getenv("POSTGRES_password"), os.Getenv("POSTGRES_user"), os.Getenv("POSTGRES_db_name")
	//создаем ссылку для подключения к бд
	DBlink := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, DBuser, DBpassword, DBname)
	//попытка открыть
	db, err := sql.Open("postgres", DBlink)
	if err != nil {
		log.Fatal("Error while opening DB", err)
	}
	//есть коннект?
	err = db.Ping()
	if err != nil {
		log.Fatal("Error while pinging DB", err)
	}
	return db
}
