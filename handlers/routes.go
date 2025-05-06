package handlers

import (
	"database/sql"

	"github.com/go-chi/chi"
)

func WalletRoutes(db *sql.DB, r *chi.Mux) {

	r.Route("/api/v1/wallets", func(r chi.Router) {
		r.Route("/{WALLET_UUID}", func(r chi.Router) {
			// GET /api/v1/wallets/{WALLET_UUID} - Получение информации о кошельке
			r.Get("/", GetWalletHandler(db))

			// POST /api/v1/wallets/{WALLET_UUID}/operations - Операции с кошельком
			r.Post("/operations", WalletOperationHandler(db))
		})
	})

}
