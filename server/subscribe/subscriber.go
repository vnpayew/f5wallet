package subscribe

import (
  	"f5wallet/server/config"
      "f5wallet/server/redis"
  "fmt"
  // "encoding/json"
  // "time"
  	"context"
  	"github.com/ethereum/go-ethereum/ethclient"
  "github.com/ethereum/go-ethereum/core/types"
  // "github.com/go-redis/redis"
  "sync"
  	"log"
    // "runtime"
)
type Subscriber struct {
   cfg *config.Config
   Name string
   SocketUrl string
   Workers *WorkerPool
   Blocks *BlockList
   Active bool
}

func NewSubscriber(cfg *config.Config, redisCache *redis.RedisPool, name string, httpUrl string,socketUrl string,blocks *BlockList)  *Subscriber {
     max_client := cfg.Webserver.MaxListenRpcConnection
     workerpool := NewWorkerPool(httpUrl,max_client,redisCache)
     //Create transaction
     subscriber :=  &Subscriber{
       cfg: cfg,
       Name: name,
       SocketUrl: socketUrl,
       Workers: workerpool,
       Blocks:blocks,
       Active: true,

     }
     return subscriber
}

func (sb *Subscriber) CheckHeader(header *types.Header){
    //Query redis
    blNumber := header.Number.String()

    fmt.Println("Subscriber:",sb.Name,"Check block: ",blNumber)
    if value, ok := sb.Blocks.Get(blNumber); ok {
         if value != sb.Name {
             fmt.Println("Call worker to get transaction from block:",blNumber)
             sb.Workers.QueryTransaction(header)
         }else{
           fmt.Println("Same subscriber received same block :",blNumber)
         }
    } else {
        fmt.Println("Not find blockNumber:",blNumber)
        sb.Blocks.Set(blNumber,sb.Name)
    }

}

func (sb *Subscriber) ListenBlockEvent(){
		fmt.Println("Subscriber:", sb.Name ,"Listening from: ", sb.SocketUrl)
		websocket, err := ethclient.Dial("ws://" + sb.SocketUrl)
		if err != nil {
				fmt.Println("Cannot connect to websocket: ", err)
				return
		}
		headers := make(chan *types.Header)
		sub, err := websocket.SubscribeNewHead(context.Background(), headers)
		if err != nil {
		    fmt.Println("Cannot SubscribeNewHead to host: ", sb.SocketUrl ," Error: ",err)
				return
		}
	  fmt.Println("Start listening: ",sb.SocketUrl,"  ")
		for {
					select {
								case err := <-sub.Err():
										fmt.Println("Error from: ",sb.SocketUrl," Error: ",err)
										log.Fatal(err)
								case header := <-headers:
                   fmt.Println("Block Number: ", header.Number.String()," Subscriber: ", sb.Name, " call CheckHeader")
                    //Process header
                    go func(){
                        sb.CheckHeader(header)
                    }()
						}
		}
}

func (sb *Subscriber) Start(wg *sync.WaitGroup){
    wg.Add(2)
    go func (){
        defer wg.Done()
        fmt.Println("Loop Subscriber waiting event: ", sb.Name)
        sb.ListenBlockEvent()
        fmt.Println("Finish Subscriber waiting event: ", sb.Name)

    }()
    go func (){
        defer wg.Done()
        fmt.Println("Loop Subscriber query transactions ", sb.Name)
        sb.Workers.LoopQueryTransaction()

        fmt.Println("Finish Subscriber query transactions: ", sb.Name)
    }()
}
