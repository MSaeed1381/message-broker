package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type Postgres struct {
	DB  *sql.DB
	Uri string
}

func NewPostgres(psqlUri string) *Postgres {
	db, err := sql.Open("postgres", psqlUri+"?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	return &Postgres{DB: db, Uri: psqlUri}
}

// Close implement closable interface
func (p *Postgres) Close() {
	err := p.DB.Close()
	if err != nil {
		panic(err)
	}
}
