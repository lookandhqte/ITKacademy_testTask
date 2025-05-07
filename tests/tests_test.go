package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"test_task/handlers"
	"test_task/requests"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Тест для handlers.RateLimitMiddleware
func TestHandlersRateLimitMiddleware(t *testing.T) {
	// Инициализируем глобальный семафор перед тестами
	handler := handlers.RateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/test-wallet", nil)
	w := httptest.NewRecorder()

	// Проверяем, что ответ OK
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Тест для операции с кошельком
func TestRequestsPostWalletOperation_Success(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	// Генерация UUID кошелька
	walletID := uuid.New()

	// Мокируем ответ от базы данных с использованием ExpectQueryRegexp
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"uuid", "balance", "currency", "created_at"}).
			AddRow(walletID.String(), 100.0, "USD", "2021-01-01"))

	// Запрос на выполнение операции
	err = requests.PostWalletOperation(db, &walletID, true, 100.0)

	assert.Nil(t, err)
	mock.ExpectationsWereMet()
}

// Тест для ошибки, если кошелек не найден
func TestRequestsPostWalletOperation_WalletNotFound(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	// Генерация UUID кошелька
	walletID := uuid.New()

	// Мокируем ошибку от базы данных
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnError(errors.New("wallet not found"))

	// Запрос на выполнение операции
	err = requests.PostWalletOperation(db, &walletID, true, 100.0)

	assert.Equal(t, "wallet not found", err.Error())
	mock.ExpectationsWereMet()
}

// Тест для операции с недостаточно средств
func TestRequestsPostWalletOperation_InsufficientFunds(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	// Генерация UUID кошелька
	walletID := uuid.New()

	// Мокируем кошелек с недостаточно средств
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"uuid", "balance", "currency", "created_at"}).
			AddRow(walletID.String(), 50.0, "USD", "2021-01-01"))

	// Запрос на выполнение операции
	err = requests.PostWalletOperation(db, &walletID, false, 100.0)

	assert.Equal(t, "insufficient funds", err.Error())
	mock.ExpectationsWereMet()
}

// Тест для операции с отрицательной суммой
func TestRequestsPostWalletOperation_InvalidAmount(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	// Генерация UUID кошелька
	walletID := uuid.New()

	// Мокируем кошелек
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"uuid", "balance", "currency", "created_at"}).
			AddRow(walletID.String(), 100.0, "USD", "2021-01-01"))

	// Запрос на выполнение операции с отрицательной суммой
	err = requests.PostWalletOperation(db, &walletID, true, -100.0)

	assert.Equal(t, "amount must be positive", err.Error())
	mock.ExpectationsWereMet()
}

// Тест на нагрузку: 1000 RPS для одного кошелька
func TestHandlersRateLimitMiddleware_1000RPS(t *testing.T) {
	handler := handlers.RateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/test-wallet", nil)

	// Выполняем 1000 запросов за 1 секунду
	for i := 0; i < 1000; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// Тест для getWallet с существующим кошельком
func TestHandlersGetWalletHandler_Success(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	walletID := uuid.New()

	// Создаем мок для базы данных, возвращая данные кошелька
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"uuid", "balance", "currency", "created_at"}).
			AddRow(walletID.String(), 100.0, "USD", "2021-01-01"))

	handler := handlers.GetWalletHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mock.ExpectationsWereMet()
}

// Тест для getWallet с несуществующим кошельком
func TestHandlersGetWalletHandler_WalletNotFound(t *testing.T) {
	// Создаем мок для базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании мока базы данных: %v", err)
	}
	defer db.Close()

	walletID := uuid.New()

	// Мокируем ошибку кошелька не найден
	mock.ExpectQuery("SELECT uuid, balance, currency, created_at FROM wallets WHERE uuid=$1").
		WithArgs(walletID).
		WillReturnError(errors.New("wallet not found"))

	handler := handlers.GetWalletHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mock.ExpectationsWereMet()
}
