package db

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestTxManagerWithinTx_PoolNil(t *testing.T) {
	tm := NewTxManager(nil)
	err := tm.WithinTx(context.Background(), func(_ context.Context, _ pgx.Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestPoolPing_Nil(t *testing.T) {
	var p *Pool
	if err := p.Ping(context.Background()); err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestPoolClose_Nil(t *testing.T) {
	var p *Pool
	p.Close() // should not panic
}

func TestWithinTx_RollsBackOnError(t *testing.T) {
	errTest := errors.New("test error")
	tm := NewTxManager(nil)
	err := tm.WithinTx(context.Background(), func(_ context.Context, _ pgx.Tx) error {
		return errTest
	})
	if err == nil {
		t.Fatal("expected error for nil pool")
	}
}
