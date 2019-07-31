package main

import (
	"f5wallet/server/config"
	"f5wallet/server/rpc"
	"f5wallet/server/redis"
	"f5wallet/server/api"
	"fmt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
  "os"
	"sync"
)


var cfg *config.Config
var rpcrouting *rpc.RpcRouting
var redisCache *redis.RedisPool
var resultCache *redis.RedisPool


func init() {
  config_file := "config.yaml"
  if len(os.Args) == 2 {
      config_file = os.Args[1]
   }

   println("init function")
   cfg = config.LoadConfig(config_file)
}

func main() {
	//Creat redis connection
	println("Initialize redis")
	redisCache = redis.NewRedisPool(cfg)

	println("Delete old data in redis ")
	//utils.DeleteData("transaction*")
	//utils.DeleteData("nonce*")

	 //Load all wallets in hosts
	 println("Create rpc connection pool ")
	 rpcrouting = rpc.NewRouting(cfg)


	 var wg sync.WaitGroup

	 wg.Add(3)

	 go func (){
				println("Loop Routing process message ")
				defer wg.Done()
				rpcrouting.Process()
	 }()

	 go func (){
		 	  println("Loop redisPool ")
			  defer wg.Done()
				redisCache.Process()
	 }()

	 go func (){
			 println("Loop httpServer ")
			defer wg.Done()
			httpServer()
	 }()

	 wg.Wait()
	 fmt.Println("Finished webserver")
}

func httpServer(){
	router := routing.New()

	//
	// api_v1 := router.Group("/api/v1/wallet")
	// fast_api := utils.NewApiFast(cfg,rpcrouting,redisCache)
  // api_v1.Get("/<method>/<p1>/<p2>/<p3>/<p4>/<p5>/<p6>", fast_api.ProcessCall)
	// api_v1.Get("/<method>/<p1>/<p2>/<p3>/<p4>/<p5>", fast_api.ProcessCall)
	// api_v1.Get("/<method>/<p1>/<p2>/<p3>/<p4>", fast_api.ProcessCall)
	// api_v1.Get("/<method>/<p1>/<p2>/<p3>", fast_api.ProcessCall)
	// api_v1.Get("/<method>/<p1>/<p2>",  fast_api.ProcessCall)
	// api_v1.Get("/<method>/<p1>", fast_api.ProcessCall)
	// api_v1.Get("/<method>",  fast_api.ProcessCall)

	api_v2 := router.Group("/api/v2/wallet")
	fast_api_v2 := api.NewApiFastV2(cfg,rpcrouting,redisCache)
	api_v2.Get("/<method>/<p1>/<p2>/<p3>/<p4>/<p5>/<p6>", fast_api_v2.ProcessCall)
	api_v2.Get("/<method>/<p1>/<p2>/<p3>/<p4>/<p5>", fast_api_v2.ProcessCall)
	api_v2.Get("/<method>/<p1>/<p2>/<p3>/<p4>", fast_api_v2.ProcessCall)
	api_v2.Get("/<method>/<p1>/<p2>/<p3>", fast_api_v2.ProcessCall)
	api_v2.Get("/<method>/<p1>/<p2>",  fast_api_v2.ProcessCall)
	api_v2.Get("/<method>/<p1>", fast_api_v2.ProcessCall)
	api_v2.Get("/<method>",  fast_api_v2.ProcessCall)

	fmt.Println("Start listening")
	panic(fasthttp.ListenAndServe(":"+ cfg.Webserver.Port, router.HandleRequest))
}
