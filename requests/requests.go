package requests

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	UUID      uuid.UUID `json:"uuid"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

func GetWallet(db *sql.DB, walletId *uuid.UUID) (*Wallet, error) {
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

func PostWalletOperation(db *sql.DB, walletId *uuid.UUID, operationType bool, amount float64) error {
	//логика изменения баланса в зависимости от типа операции
	wallet, err := GetWallet(db, walletId)
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
