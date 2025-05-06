package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error while loading .env file", err)
	}

	const (
		host = "localhost"
		port = 5432
	)
	//получаем пароль, имя пользователя и имя бд из окружающего конфиг.env файла
	DBpassword, DBuser, DBname := os.Getenv("POSTGRES_password"), os.Getenv("POSTGRES_user"), os.Getenv("POSTGRES_db_name")
	DBlink := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, DBuser, DBpassword, DBname)
	db, err := sql.Open("postgres", DBlink)
	if err != nil {
		log.Fatal("Error while opening DB", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error while pinging DB", err)
	}
	defer db.Close()
	fmt.Println("Successfully connected!")
}
