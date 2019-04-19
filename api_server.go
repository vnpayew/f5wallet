package main

import (
	"f5wallet/server/config"
	"f5wallet/server/redis"
	"f5wallet/server/rpc"
	"f5wallet/server/api"
	"f5wallet/server/jwt"
	"fmt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
  "os"
  // "strings"
  // "github.com/go-redis/redis"
  // "encoding/json"
	// "github.com/savsgio/go-logger"
	"sync"
)


var cfg *config.Config
var rpcrouting *rpc.RpcRouting
var redisCache *redis.RedisPool

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

	 //Load all wallets in hosts
	 println("Create rpc connection pool ")
	 rpcrouting = rpc.NewRouting(cfg)

	 // println("Load wallets ")
 	 // account.LoadWallets(rpcrouting)

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
			ApiServer()
	 }()

	 wg.Wait()
	 fmt.Println("Finished webserver")
}

func ApiServer(){

	// router := fasthttprouter.New()
	router := routing.New()

	ap := api.NewF5ApiV1(cfg,rpcrouting, redisCache)
  router.Get("/login", ap.Login_get)
	router.Post("/login", ap.Login)

	api := router.Group("/api/v1")

	if cfg.Jwt.Enable {
			fmt.Println("Using jwt")
	}

	api.Put("/cash/credit/<address>/<amount>/<traceid>", jwt.JWTMiddleware(ap.CashCredit))
	api.Put("/cash/debit/<address>/<amount>/<traceid>", jwt.JWTMiddleware(ap.CashDebit))
	api.Put("/cash/transfer/<from>/<to>/<amount>/<note>/<traceid>", jwt.JWTMiddleware(ap.CashTransfer))


	api.Get("/balance/<address>", jwt.JWTMiddleware(ap.Balance))
	api.Get("/balance/all", jwt.JWTMiddleware(ap.BalanceAll))

	api.Post("/account/new", jwt.JWTMiddleware(ap.AccountNew))
	api.Get("/account/total", jwt.JWTMiddleware(ap.AccountTotal))
	api.Get("/account/list/active", jwt.JWTMiddleware(ap.AccountListActive))
	api.Get("/account/list/inactive", jwt.JWTMiddleware(ap.AccountListInactive))
	api.Get("/account/lock/<address>/<traceid>", jwt.JWTMiddleware(ap.AccountLock))
	api.Get("/account/status/<address>", jwt.JWTMiddleware(ap.AccountStatus))

	api.Get("/transaction/<txhash>",jwt.JWTMiddleware(ap.Transaction))
	api.Get("/transaction/list/<account>/<fromdate>/<todate>", jwt.JWTMiddleware(ap.TransactionList))
	api.Get("/transaction/lock/<account>/<fromdate>/<todate>", jwt.JWTMiddleware(ap.TransactionLock))

	server := &fasthttp.Server{
		Name:    "JWT API Server",
		Handler: router.HandleRequest,
	}

	fmt.Println("Start listening")

	if cfg.Webserver.Tls {
		fmt.Println("Start server using TLS ")
		panic(server.ListenAndServeTLS(":"+ cfg.Webserver.Port,cfg.Webserver.CertificateFile,cfg.Webserver.KeyFile))
	} else {
		fmt.Println("Start server without TLS  ")
		panic(server.ListenAndServe(":"+ cfg.Webserver.Port))
	}
}
