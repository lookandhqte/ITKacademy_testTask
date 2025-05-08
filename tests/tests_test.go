package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"test_task/config"
	"test_task/handlers"
	"test_task/requests"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func loadEnv() error {
	err := godotenv.Load("config.env")
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}
func TestGetWalletHandler(t *testing.T) {
	// Подключаемся к реальной базе данных
	loadEnv()
	db := config.ConnectToDB()
	defer db.Close()

	// Генерация нового UUID
	walletUUID := uuid.New()

	// Создаем тестовую строку для запроса
	req, err := http.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletUUID.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем запись для кошелька в базе данных
	_, err = db.Exec(`INSERT INTO wallets (uuid, balance, currency, created_at) VALUES ($1, $2, $3, $4)`,
		walletUUID, 1000.0, "USD", "2025-01-01")
	if err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	handlers.WalletRoutes(db, r)
	ts := httptest.NewServer(r) // Создаем сервер для тестов
	defer ts.Close()

	// Создаем сервер и записываем запрос
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)

	var wallet requests.Wallet
	if err := json.NewDecoder(rr.Body).Decode(&wallet); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, walletUUID, wallet.UUID)
}

func TestPostWalletOperationHandler(t *testing.T) {
	// Подключаемся к реальной базе данных
	loadEnv()
	db := config.ConnectToDB()
	defer db.Close()

	// Генерация нового UUID для кошелька
	walletUUID := uuid.New()

	// Создаем тестовую строку для запроса
	requestBody := `{"uuid":"` + walletUUID.String() + `", "operation_type": true, "amount": 100.0}`
	req, err := http.NewRequest(http.MethodPost, "/api/v1/wallets/"+walletUUID.String()+"/operations", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем запись для кошелька в базе данных
	_, err = db.Exec(`INSERT INTO wallets (uuid, balance, currency, created_at) VALUES ($1, $2, $3, $4)`,
		walletUUID, 1000.0, "USD", "2025-01-01")
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	handlers.WalletRoutes(db, r)
	ts := httptest.NewServer(r) // Создаем сервер для тестов
	defer ts.Close()

	// Создаем сервер и записываем запрос
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Проверяем статус ответа
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем обновленный баланс кошелька
	var balance float64
	err = db.QueryRow(`SELECT balance FROM wallets WHERE uuid = $1`, walletUUID).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1100.0, balance)
}

func TestRateLimitMiddleware(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/wallets/some-wallet", nil)
	if err != nil {
		t.Fatal(err)
	}
	loadEnv()

	db := config.ConnectToDB()
	defer db.Close()

	handler := handlers.RateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
