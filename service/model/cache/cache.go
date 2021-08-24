package cache

import (
	"github.com/go-redis/redis"
	"log"
)

var Client *redis.Client

func init() {
	log.Printf("cache init")
	Client = redis.NewClient(&redis.Options{
		Addr:     "172.17.0.1:6379",
		Password: "Wwcwwc123",
		DB:       0,
	})
	log.Printf("cache init succeed")
}

func Publish(channel string, msg string) error {
	_, err := Client.Publish(channel, msg).Result()
	return err
}
