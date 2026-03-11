package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rithul/hivemind/registry/api/internal/auth"
	"github.com/rithul/hivemind/registry/api/internal/config"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := db.New(pool)

	// Create a user
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        fmt.Sprintf("test%d@example.com", time.Now().Unix()),
		Username:     fmt.Sprintf("testuser_%d", time.Now().Unix()),
		PasswordHash: pgtype.Text{String: "hash", Valid: true},
	})
	if err != nil {
		log.Fatal("CreateUser:", err)
	}

	// Create an API key
	rawKey := "hm_test_key_12345"
	hash := auth.HashKey(rawKey)

	key, err := queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:    user.ID,
		Name:      "test_key",
		KeyHash:   hash,
		KeyPrefix: "hm_test_",
		Scopes:    []string{"publish"},
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		log.Fatal("CreateAPIKey:", err)
	}

	fmt.Printf("API_KEY=%s\n", rawKey)
	fmt.Printf("USER_ID=%s\n", user.ID.String())
	fmt.Printf("KEY_ID=%s\n", key.ID.String())
}
