package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VectorDB struct {
	Dsn    string        `json:"dsn"` // 数据库连接字符串
	Dbpool *pgxpool.Pool `json:"-"`   // PG Vector数据库连接池
}

// DSN 数据库连接字符串 like "postgres://user:passwd@localhost:5432/ragdb?sslmode=disable"
func (p *VectorDB) Open(dsn string) error {
	p.Dsn = dsn
	dbpool, err := pgxpool.New(context.Background(), p.Dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return err
	}

	p.Dbpool = dbpool

	return nil
}

func (p *VectorDB) Close() error {
	if (p.Dbpool) != nil {
		p.Dbpool.Close()
		p.Dbpool = nil
	}
	return nil
}
