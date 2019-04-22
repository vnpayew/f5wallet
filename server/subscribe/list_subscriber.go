package subscribe

import (
  	"f5wallet/server/config"
    "f5wallet/server/redis"
  "fmt"
  // "encoding/json"
  "time"
  	// "context"
  // 	"github.com/ethereum/go-ethereum/ethclient"
  "github.com/ethereum/go-ethereum/core/types"
  // "github.com/go-redis/redis"
  "sync"
  	// "log"
    "context"
    "strings"
)



type SubscriberPool struct {
    cfg *config.Config
    Subscribers []*Subscriber
    Workers []*WorkerPool
    Blocks *BlockList
    PolicyLevel int
    redisCache *redis.RedisPool    //Redis pool
    Queue *RabbitMQProducer
    mutex sync.Mutex
    wg *sync.WaitGroup
}

var subpool *SubscriberPool

func isConnectionError(err error) bool {
    err_msg := err.Error()
    if strings.Contains(err_msg, "connection refused") {
        return true
    }
    return false
}

func NewSubscriberPool(cfg *config.Config, redisCache *redis.RedisPool) *SubscriberPool{
    max_client := cfg.Webserver.MaxListenRpcConnection

    var subscribers []*Subscriber
    var workers []*WorkerPool
    for _,host := range cfg.Networks {
        sb := NewSubscriber(host.Name,host.WebSocket)
        subscribers = append(subscribers,sb)

        worker := NewWorkerPool(host.Name,host.Http,max_client)
        workers = append(workers,worker)
    }

    blockList := NewBlockList()

    policyLevel := len(subscribers)

    queue := NewRabbitMQProducer(cfg.RabbitMq.Url, cfg.RabbitMq.QueueName, cfg.RabbitMq.MaxClient )

    subpool := &SubscriberPool{
      cfg: cfg,
      Subscribers: subscribers,
      Workers: workers,
      Blocks: blockList,
      redisCache: redisCache,
      PolicyLevel: policyLevel,
      Queue: queue,
    }

    //Set parent for each subscriber
    for _,sub := range subscribers{
       sub.SetParent(subpool)
    }
    return subpool
}

func (sp *SubscriberPool) Start(wg *sync.WaitGroup){
    //Start all subscriber
    for i, sub := range sp.Subscribers {
        fmt.Println("Start subscriber: ",sub.Name," index: ", i)
        wg.Add(1)
        go func (s *Subscriber){
              defer wg.Done()
              fmt.Println("Loop Subscriber waiting event: ", s.Name," index: ", i)
              s.ListenBlockEvent()
              fmt.Println("Finish Subscriber waiting event: ", s.Name," index: ", i)
        }(sub)
    }
    sp.wg = wg
    //Start process to check health
    wg.Add(1)
    go func (){
        defer wg.Done()
        fmt.Println("Start loop to check health ")
        sp.CheckHealth()
        fmt.Println("End loop check heath ")
    }()
}

func (sp *SubscriberPool) CheckHealth(){
      for {
          select {
          case <-time.After(5*time.Second):
               go func() {
                 sp.CheckSubscriberHealth()
               }()
          case <-time.After(10*time.Second):
              go func() {
                sp.CheckWorkerHealth()
              }()
        }
      }
}
func (sp *SubscriberPool) CheckSubscriberHealth() {
    //fmt.Println("Process to check subscriber connections to node. To automatically active disabled node")
    for _, sub := range sp.Subscribers {
        if !sub.Active {
            fmt.Println("Retry listen block event from host: ", sub.Name)
            sp.wg.Add(1)
            go func (s *Subscriber){
                defer sp.wg.Done()
                fmt.Println("Try to enable subscriber: ", s.Name)
                s.ListenBlockEvent()
              }(sub)
        }
     }
}
func (sp *SubscriberPool) CheckWorkerHealth() {
    //fmt.Println("Process to check worker connections to node. To automatically active disabled node")
    for _, worker := range sp.Workers {
      if !worker.Active {
          fmt.Println("Retry connect to host: ", worker.Name)
          worker.UpdateHealth()
      }
   }
}
func (sp *SubscriberPool) GetPolicyLevel() int {
    sp.mutex.Lock()
    defer sp.mutex.Unlock()
    return sp.PolicyLevel
}
func (sp *SubscriberPool) GetActiveWorker(names []string) *WorkerPool{
  sp.mutex.Lock()
  defer sp.mutex.Unlock()
  for _, name := range names {
      for _, worker := range sp.Workers {
          if (worker.Name == name && worker.Active) {
              return worker
          }
      }
  }
  return nil
}
func (sp *SubscriberPool) UpdatePolicyLevel(){
  count := 0;
  for _, sub := range sp.Subscribers {
    if (sub.Active) {
      count  = count + 1
    }
  }
  sp.mutex.Lock()
  defer sp.mutex.Unlock()
  sp.PolicyLevel = count
}
func (sp *SubscriberPool) CheckPolicyBlock(header *types.Header) {
    blNumber := header.Number.String()
    if values, ok := sp.Blocks.Get(blNumber); ok {
        level := sp.GetPolicyLevel()
        if (level >0 ){
              if(len(values) >= level) {
                   fmt.Println("Send message to worker to query transaction")
                   sp.QueryTransaction(values, header)
              } else {
                 fmt.Println("Not enough info. Wait other subscriber")
              }
        } else {
           fmt.Println(" No active subscribers. Error in new block event")
        }
    } else {
       fmt.Println(" Add Block: ", blNumber," Error before")
    }
}

func (sp *SubscriberPool) UpdateBlockHeader(name string, header *types.Header){
    sp.mutex.Lock()
    defer sp.mutex.Unlock()

    sp.Blocks.AddBlock(header,name)

    //Check Policy thread
    go func (){
          fmt.Println("New thread to made decision for block: ", header.Number.String())
          sp.CheckPolicyBlock(header)
    }()
}

func (sp *SubscriberPool) QueryTransaction(names []string, header *types.Header ){
    retry := 0
    for retry < 3 {
          worker := sp.GetActiveWorker(names)
          if worker == nil {
                fmt.Println("Cannot find worker to query transaction ")
                return
          }

          conn := worker.GetConnection()

          conn.Mux.Lock()
          defer  conn.Mux.Unlock()

          fmt.Println("Try to query block hash: ",header.Number.String() )
          block, err := conn.Client.BlockByHash(context.Background(), header.Hash())
          if err == nil {
                t := time.Now()
                fmt.Println(t.Format(time.RFC822),"Host:", worker.Name, ", Block Number: ", header.Number.String(),"number of transactions:", len(block.Transactions()), " header hash: " , header.Hash().Hex())
                coinbase := block.Coinbase()
                fmt.Println("Start write ",len(block.Transactions())," of transaction of block: ",header.Number.String()," to redis" )
                for _, transaction := range block.Transactions(){
                     nonce := transaction.Nonce()
                     key := strings.TrimPrefix(transaction.Hash().Hex(),"0x")

                     sp.redisCache.LogEnd(key,nonce,coinbase.Hex())

                     bs, err := transaction.MarshalJSON()
                     if err != nil {
                         fmt.Println("Error to convert transaction to json: ", err)
                         continue
                     }
                     fmt.Println(t.Format(time.RFC822),"Store: ",string(bs)," to redis")
                     sp.Queue.Publish(string(bs))
                }
                fmt.Println("Host:", worker.Name, ",Finish writing ",len(block.Transactions())," of transaction of block: ",header.Number.String()," to redis" )
                return
          }
          if isConnectionError(err) {
             fmt.Println("Error in network to get block: ",err,". Disable worker: ", worker.Name)
             worker.Active = false
          } else {
             fmt.Println("Other error: ",err)
             return
          }
          retry = retry + 1
          fmt.Println("Network error. Retry: ", retry)
      }
}
