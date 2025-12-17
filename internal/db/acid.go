package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type transactionKey string

var txKey transactionKey = "key"

type acid struct {
	db *sqlx.DB
}

func (ac acid) WithTransaction(ctx context.Context) (context.Context, error) {
	_, alreadyMarked := ac.txFromContext(ctx)
	if alreadyMarked {
		return ctx, nil
	}
	tx, err := ac.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	ctx = ac.newTxContext(ctx, tx)
	return ctx, nil
}

func (ac acid) Save(ctx context.Context) error {
	tx, ok := ac.txFromContext(ctx)
	// no-op
	if !ok {
		return nil
	}
	defer ac.removeTx(ctx)
	err := tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ac acid) Rollback(ctx context.Context) error {
	tx, ok := ac.txFromContext(ctx)
	// no-op
	if !ok {
		return nil
	}
	defer ac.removeTx(ctx)
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}

func (ac acid) newTxContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	// using **sqlx.Tx allows for the value to be deleted(set to nil)
	return context.WithValue(ctx, txKey, &tx)
}

func (ac acid) txFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	v, ok := ctx.Value(txKey).(**sqlx.Tx)
	if !ok {
		return nil, ok
	}
	ok = *v != nil

	return *v, ok
}

func (ac acid) beginTxx(ctx context.Context, db *sqlx.DB) (*sqlx.Tx, error) {
	tx, existingTx := ac.txFromContext(ctx)
	if !existingTx {
		newTx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return nil, err
		}
		tx = newTx
	}
	return tx, nil
}

func (ac acid) rollbackTxx(ctx context.Context, tx *sqlx.Tx) {
	// check it the tx came from ctx, if it is no-op.
	tmpTx, ok := ac.txFromContext(ctx)
	if ok && tmpTx == tx {
		return
	}
	rollbackTxx(tx)
}

func (ac acid) commitTxx(ctx context.Context, tx *sqlx.Tx) error {
	// check it the tx came from ctx, if it is no-op.
	tmpTx, ok := ac.txFromContext(ctx)
	if ok && tmpTx == tx {
		return nil
	}
	return tx.Commit()
}

// getDbtx will give dbtx based on the ctx. if context is marked for transaction, it will return the tx. if not, will return the db instead.
func (ac acid) getDbtx(ctx context.Context, db *sqlx.DB) dbtx {
	tx, ok := ac.txFromContext(ctx)
	if ok {
		return tx
	}
	return db
}

func (ac acid) removeTx(ctx context.Context) {
	tx, ok := ctx.Value(txKey).(**sqlx.Tx)
	if !ok {
		return
	}
	*tx = nil
}
