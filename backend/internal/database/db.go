package database

import "github.com/jackc/pgx/v5/pgxpool"

var pool *pgxpool.Pool

func SetPool(p *pgxpool.Pool) {
	pool = p
}

func GetPool() *pgxpool.Pool {
	return pool
}
