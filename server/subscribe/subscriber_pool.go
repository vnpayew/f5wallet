package subscribe

import (
  	"f5wallet/server/config"
      "f5wallet/server/redis"
  "fmt"
  // "encoding/json"
  // "time"
  	// "context"
  // 	"github.com/ethereum/go-ethereum/ethclient"
  // "github.com/ethereum/go-ethereum/core/types"
  // "github.com/go-redis/redis"
  "sync"
  	// "log"
)



type SubscriberPool struct {
    cfg *config.Config
    Subscribers []*Subscriber
    Blocks *BlockList
}

var subpool *SubscriberPool

func NewSubscriberPool(cfg *config.Config, redisCache *redis.RedisPool) *SubscriberPool{
    blockList := NewBlockList()

    var subscribers []*Subscriber
    for _,host := range cfg.Networks {
        sb := NewSubscriber(cfg, redisCache, host.Name,host.Http,host.WebSocket,blockList)
        subscribers = append(subscribers,sb)
    }

    subpool := &SubscriberPool{
      cfg: cfg,
      Subscribers:subscribers,
      Blocks: blockList,
    }
    return subpool
}

func (sp *SubscriberPool) Start(wg *sync.WaitGroup){
    for _,sub := range sp.Subscribers {
      fmt.Println("Start subscriber: ",sub.Name)
      sub.Start(wg)
    }
}
