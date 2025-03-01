package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"

	"github.com/snowmerak/keycl/lib/store/queries"
)

type Store struct {
	conn *pgx.Conn
	lock sync.Mutex
}

func New(ctx context.Context, connectionString string) (*Store, error) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("pgx.Connect: %w", err)
	}

	context.AfterFunc(ctx, func() {
		conn.Close(ctx)
	})

	return &Store{conn: conn}, nil
}

func (s *Store) Visit(ctx context.Context, visitor func(ctx context.Context, q *queries.Queries) error) error {
	q := queries.New(s.conn)
	return visitor(ctx, q)
}
