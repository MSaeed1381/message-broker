package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"runtime"
	"sync"
)

type Postgres struct {
	db *pgxpool.Pool
}

// singleton design pattern with once keyword
// singleton pattern to make sure that I only have one connection pool
var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string) (*Postgres, error) {
	pgOnce.Do(func() {
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			panic(err)
		}

		config.MaxConns = int32(runtime.NumCPU())
		config.MinConns = 2

		db, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			panic(err)
		}

		pgInstance = &Postgres{db: db}
	})

	fmt.Println("connected to postgres...")
	return pgInstance, nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}

// Close implement closable interface
func (p *Postgres) Close() {
	p.db.Close()
}
