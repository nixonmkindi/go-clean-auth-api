package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required")
	}

	direction := "up"
	if len(os.Args) > 1 {
		direction = strings.ToLower(os.Args[1])
	}
	if direction != "up" && direction != "down" {
		panic("usage: go run cmd/migrate/main.go [up|down]")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	pattern := filepath.Join("migrations", fmt.Sprintf("*.%s.sql", direction))
	files, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		fmt.Println("no migration files found")
		return
	}

	sort.Strings(files)
	if direction == "down" {
		reverse(files)
	}

	for _, f := range files {
		body, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		if strings.TrimSpace(string(body)) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			panic(fmt.Sprintf("migration failed: %s: %v", f, err))
		}
		fmt.Printf("applied %s\n", f)
	}
}

func reverse(items []string) {
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
}
