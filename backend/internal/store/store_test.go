package store

import (
	"context"
	"testing"
	"time"
)

func TestOpenRejectsInvalidURL(t *testing.T) {
	if _, err := Open(context.Background(), "postgres://bad url"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestOpenGivesUpWhenContextEnds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	_, err := Open(ctx, "postgres://u:p@127.0.0.1:1/db?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("expected connection error")
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("Open took %v, want it to stop at the context deadline", d)
	}
}
