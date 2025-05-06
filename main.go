package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	host = "localhost"
	port = 5432
)

func loadEnv() error {
	err := godotenv.Load("config.env")
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}

func connectToDB() *sql.DB {
	DBpassword, DBuser, DBname := os.Getenv("POSTGRES_password"), os.Getenv("POSTGRES_user"), os.Getenv("POSTGRES_db_name")
	//создаем ссылку для подключения к бд
	DBlink := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, DBuser, DBpassword, DBname)
	//попытка открыть
	db, err := sql.Open("postgres", DBlink)
	if err != nil {
		log.Fatal("Error while opening DB", err)
	}
	//попытка пропинговать
	err = db.Ping()
	if err != nil {
		log.Fatal("Error while pinging DB", err)
	}
	//закрытие базы данных
	fmt.Println("Successfully connected!")
	return db
}

func main() {
	//подгружаем данные из .env файла
	loadEnv()
	db := connectToDB()
	defer db.Close()
	r := chi.NewRouter()
	walletRoutes(db, r)

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}

func walletRoutes(db *sql.DB, r *chi.Mux) {

	r.Route("/api/v1/wallets", func(r chi.Router) {
		r.Route("/{WALLET_UUID}", func(r chi.Router) {
			// GET /api/v1/wallets/{WALLET_UUID} - Получение информации о кошельке
			r.Get("/", GetWalletHandler(db))

			// POST /api/v1/wallets/{WALLET_UUID}/operations - Операции с кошельком
			r.Post("/operations", WalletOperationHandler(db))
		})
	})

}

func GetWalletHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		walletId, err := uuid.Parse(chi.URLParam(r, "WALLET_UUID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		wallet, err := getWallet(db, &walletId)
		if err != nil {
			if err.Error() == "wallet not found" {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wallet)
	}
}

func WalletOperationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		walletId, err := uuid.Parse(chi.URLParam(r, "WALLET_UUID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		var req WalletOperationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := postWalletOperation(db, &walletId, req.OperationType, req.Amount); err != nil {
			switch err.Error() {
			case "wallet not found":
				http.Error(w, err.Error(), http.StatusNotFound)
			case "insufficient funds", "amount must be positive":
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type Wallet struct {
	UUID      uuid.UUID `json:"uuid"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

// НЕ ИСПОЛЬЗУЕТСЯ МБ УБРАТЬ
type WalletOperationRequest struct {
	UUID          uuid.UUID `json:"uuid"`
	OperationType bool      `json:"operation_type"`
	Amount        float64   `json:"amount"`
}

func getWallet(db *sql.DB, walletId *uuid.UUID) (*Wallet, error) {
	newWallet := &Wallet{}
	err := db.QueryRow(`SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1`, walletId).Scan(&newWallet.UUID, &newWallet.Balance, &newWallet.Currency, &newWallet.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return newWallet, nil
}

func postWalletOperation(db *sql.DB, walletId *uuid.UUID, operationType bool, amount float64) error {
	//логика изменения баланса в зависимости от типа операции
	wallet, err := getWallet(db, walletId)
	//ошибка получения кошелька
	if err != nil {
		return err
	}

	//валидация суммы
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	//транзакция
	tx, err := db.Begin()
	//ошибка в транзакции
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if !operationType { //WITHDRAW
		if wallet.Balance < amount {
			return errors.New("insufficient funds")
		}
		wallet.Balance -= amount
	} else { //DEPOSIT
		wallet.Balance += amount
	}
	_, err = tx.Exec(`UPDATE wallets SET balance = $1 WHERE uuid = $2`, wallet.Balance, walletId)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
