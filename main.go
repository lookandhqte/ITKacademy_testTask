package main

import (
	"fmt"
	"log"
	"net/http"
	"test_task/config"
	"test_task/handlers"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func loadEnv() error {
	err := godotenv.Load("config.env")
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}

func main() {
	//подгружаем данные из .env файла
	loadEnv()
	db := config.ConnectToDB()
	fmt.Println("Successfully connected!")
	defer db.Close()
	r := chi.NewRouter()
	handlers.WalletRoutes(db, r)

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}
