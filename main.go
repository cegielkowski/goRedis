package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	ctx := context.Background()

	redis := newRedis()

	for i := 1; i < 1000000; i++ {
		key := uuid.New().String()
		redis.Add(ctx, key+strconv.Itoa(rand.Intn(100)), "test", rand.Intn(10)*1000)
	}

	time.Sleep(10 * time.Second)
	fmt.Println(redis.ShowAll())
}
