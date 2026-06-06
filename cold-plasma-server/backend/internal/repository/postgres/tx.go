package postgres

import (
	"context"
	"fmt"

	"cold-plasma-server/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context, repos repository.TxRepos) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	repos := &txRepos{tx: tx}
	if err := fn(ctx, repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type txRepos struct {
	tx pgx.Tx
}

func (r *txRepos) User() repository.UserRepository {
	return NewUserRepo(r.tx)
}

func (r *txRepos) Booking() repository.BookingRepository {
	return NewBookingRepo(r.tx)
}

func (r *txRepos) Bonus() repository.BonusRepository {
	return NewBonusRepo(r.tx)
}

