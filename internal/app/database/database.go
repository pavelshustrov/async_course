package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

func ConnectDefault() *pgx.Conn {
	return connect("DATABASE_URL")
}

func ConnectByName(name string) *pgx.Conn {
	return connect(fmt.Sprintf("DATABASE_%s_URL", strings.ToUpper(name)))
}

func connect(name string) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), os.Getenv(name))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return conn
}
