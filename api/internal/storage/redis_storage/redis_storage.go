package redis_storage

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	redisClient *redis.Client
}

func MustConnect(addr string, pass string, db int) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass, // no password set
		DB:       db,   // use default DB
	})
	service := &RedisService{rdb}

	if err := service.Ping(); err != nil {
		panic(err)
	}
	return service
}

func (s *RedisService) Ping() error {
	return s.redisClient.Ping(context.Background()).Err()
}

func (s *RedisService) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return s.redisClient.Subscribe(ctx, channel)
}

func (s *RedisService) Publish(ctx context.Context, channel string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return s.redisClient.Publish(ctx, channel, payload).Err()
}
