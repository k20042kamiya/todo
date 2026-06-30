package database

import (
	"context"
	"log/slog"

	"todo/shared/transaction"

	"gorm.io/gorm"
)

type txKey struct{}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) transaction.Manager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			slog.ErrorContext(ctx, "transaction rollback failed", "rollback_error", rbErr, "original_error", err)
		}
		return err
	}

	return tx.Commit().Error
}

func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
