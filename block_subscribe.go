package main

import (
	"f5wallet/server/config"
	"f5wallet/server/redis"
	"f5wallet/server/subscribe"
	"os"
	// "strings"
	"fmt"
	// "github.com/ethereum/go-ethereum/core/types"
	// "github.com/go-redis/redis"
	"sync"
	// "time"
)

var cfg *config.Config


func main() {
	 var wg sync.WaitGroup

	 config_file := "config.yaml"
	 if len(os.Args) == 2 {
			config_file = os.Args[1]
	 }

	 	cfg = config.LoadConfig(config_file)

		//Creat redis Poool
		redisPool := redis.NewRedisPool(cfg)
		wg.Add(1)
		//Loop waiting to process log
		go func (){
				fmt.Println("Loop Redis Pool ")
				defer wg.Done()
				redisPool.Process()
		}()

		subPool := subscribe.NewSubscriberPool(cfg,redisPool)
		subPool.Start(&wg)

	  wg.Wait()
}
