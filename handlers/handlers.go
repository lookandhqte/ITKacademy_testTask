package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"test_task/requests"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type WalletOperationRequest struct {
	UUID          uuid.UUID `json:"uuid"`
	OperationType bool      `json:"operation_type"`
	Amount        float64   `json:"amount"`
}

func GetWalletHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		walletId, err := uuid.Parse(chi.URLParam(r, "WALLET_UUID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		wallet, err := requests.GetWallet(db, &walletId)
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
			fmt.Println(walletId)
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		var req WalletOperationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := requests.PostWalletOperation(db, &walletId, req.OperationType, req.Amount); err != nil {
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
