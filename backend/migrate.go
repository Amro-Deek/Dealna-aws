package main

import (
	"context"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	_, err = pool.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS public.provider_pre_registration (
        id uuid DEFAULT gen_random_uuid() NOT NULL,
        email character varying(255) NOT NULL,
        token uuid NOT NULL,
        expires_at timestamp without time zone NOT NULL,
        used_at timestamp without time zone,
        resend_count integer DEFAULT 0 NOT NULL,
        resend_window_start timestamp without time zone,
        created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
        verified_at timestamp without time zone,
        UNIQUE(email),
        UNIQUE(token)
    );
    ALTER TABLE public.provider_pre_registration OWNER TO dealna_user;
	`)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration successful")
}
