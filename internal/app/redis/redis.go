package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

func New(host, port string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // нет пароля
		DB:       0,  // используем БД по умолчанию
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// AddToBlacklist добавляет токен в blacklist
func (c *Client) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	return c.client.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsInBlacklist проверяет есть ли токен в blacklist
func (c *Client) IsInBlacklist(ctx context.Context, token string) (bool, error) {
	_, err := c.client.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil // токена нет в blacklist
	}
	if err != nil {
		return false, err // ошибка Redis
	}
	return true, nil // токен в blacklist
}
