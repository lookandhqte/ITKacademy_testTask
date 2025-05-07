package handlers

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"golang.org/x/sync/semaphore"
)

var (
	semaphores    sync.Map        // key: walletUUID (string), value: *semaphore.Weighted
	maxConcurrent int64    = 1000 // Максимальное число одновременных запросов
)

// Middleware для ограничения RPS
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем UUID из URL
		walletID := chi.URLParam(r, "WALLET_UUID")

		// Загружаем или создаем семафор для кошелька
		sem, _ := semaphores.LoadOrStore(walletID, semaphore.NewWeighted(maxConcurrent))

		// Пытаемся занять слот
		if err := sem.(*semaphore.Weighted).Acquire(r.Context(), 1); err != nil {
			http.Error(w, "Too many requests", http.StatusServiceUnavailable)
			return
		}
		defer sem.(*semaphore.Weighted).Release(1) // Освобождаем слот

		next.ServeHTTP(w, r)
	}
}

func WalletRoutes(db *sql.DB, r *chi.Mux) {

	r.Route("/api/v1/wallets", func(r chi.Router) {
		r.Route("/{WALLET_UUID}", func(r chi.Router) {
			// GET /api/v1/wallets/{WALLET_UUID} - Получение информации о кошельке
			r.Get("/", RateLimitMiddleware(GetWalletHandler(db)))

			// POST /api/v1/wallets/{WALLET_UUID}/operations - Операции с кошельком
			r.Post("/operations", RateLimitMiddleware(WalletOperationHandler(db)))
		})
	})

}
