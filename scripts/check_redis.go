package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	defer client.Close()

	// 检查队列长度
	length, _ := client.LLen(context.Background(), "queue:yun_express_webhook_track").Result()
	fmt.Printf("Queue length: %d\n\n", length)

	// 检查缓存中的数据
	keys := []string{
		"tracking:detail:YT2612500701966827",
		"tracking:detail:YT2612500702088888",
	}

	for _, key := range keys {
		val, err := client.Get(context.Background(), key).Result()
		if err != nil {
			fmt.Printf("Key %s: not found\n", key)
		} else {
			fmt.Printf("Key %s:\n%s\n\n", key, val)
		}
	}
}
