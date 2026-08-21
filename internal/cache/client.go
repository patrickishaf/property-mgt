package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/patrickishaf/property-mgt/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	mu     *sync.Mutex
}

func CreateClient(cfg *config.RedisConfig) *Client {
	var mu sync.Mutex
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis. err: %s", err.Error())
	}
	return &Client{
		client: redisClient,
		mu:     &mu,
	}
}

func (this *Client) Client() *redis.Client {
	return this.client
}

func (this *Client) Delete(ctx context.Context, key string) error {
	_, err := this.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (this *Client) Get(ctx context.Context, key string) (string, error) {
	return this.client.Get(ctx, key).Result()
}

func (this *Client) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.client.Set(ctx, key, value, time.Duration(ttlSeconds*int(time.Second))).Err()
}
