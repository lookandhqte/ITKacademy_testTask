package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	DBpassword, DBuser, DBname, HOST, PORT := os.Getenv("POSTGRES_password"), os.Getenv("POSTGRES_user"), os.Getenv("POSTGRES_db_name"), os.Getenv("HOST"), os.Getenv("PORT")
	//создаем ссылку для подключения к бд
	DBlink := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", HOST, PORT, DBuser, DBpassword, DBname)
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
