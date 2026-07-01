package config

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() *redis.Client {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		host := strings.TrimSpace(os.Getenv("REDIS_HOST"))
		port := strings.TrimSpace(os.Getenv("REDIS_PORT"))
		if host != "" && port != "" {
			addr = host + ":" + port
		}
	}

	if addr == "" {
		log.Println("Redis disabled: set REDIS_ADDR or REDIS_HOST + REDIS_PORT to enable Redis Cloud")
		return nil
	}

	db := 0
	if rawDB := strings.TrimSpace(os.Getenv("REDIS_DB")); rawDB != "" {
		parsed, err := strconv.Atoi(rawDB)
		if err != nil {
			log.Fatalf("invalid REDIS_DB: %v", err)
		}
		db = parsed
	}

	opt := &redis.Options{
		Addr:         addr,
		Username:     os.Getenv("REDIS_USERNAME"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           db,
		DialTimeout:  8 * time.Second,
		ReadTimeout:  8 * time.Second,
		WriteTimeout: 8 * time.Second,
	}

	if strings.EqualFold(os.Getenv("REDIS_TLS"), "true") || strings.EqualFold(os.Getenv("REDIS_TLS"), "1") {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect Redis: %v", err)
	}

	Redis = client
	log.Printf("Redis connected: %s", addr)
	return client
}
