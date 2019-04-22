package subscribe

import (
  "f5wallet/server/rpc"
  "fmt"
      "context"
  // "encoding/json"
  // "time"
  // "github.com/ethereum/go-ethereum/ethclient"
  // "github.com/ethereum/go-ethereum/core/types"
  // "github.com/go-redis/redis"
  "sync"
  	"log"
)


type WorkerPool struct {
   Name string    //Name of host
   Active bool    //Worker pool state
   HttpUrl string
   Clients []*rpc.RpcConnection   //Current connection to host
   Current int
   mutex sync.Mutex
}



func NewWorkerPool(name string, httpUrl string, max_client int)  *WorkerPool {
    //Create RPC connections
    var clients  []*rpc.RpcConnection
    for i:=0; i< max_client; i++ {
       ethclient, err := rpc.NewRpcConnection(name,httpUrl)
       if err != nil {
           log.Fatal("Cannot connect to: ",httpUrl," error:", err)
           continue
       }
       clients = append(clients,ethclient)
     }

     //Create transaction
     workerpool :=  &WorkerPool{
          Name: name,
          Active: true,
          HttpUrl: httpUrl,
          Clients: clients,
          Current: 0,
     }
     return workerpool
}

func (wp *WorkerPool) GetConnection() *rpc.RpcConnection {
  wp.mutex.Lock()
  defer wp.mutex.Unlock()

  len := len(wp.Clients)
  if wp.Current >=  len {
      wp.Current = wp.Current % len
  }
  client := wp.Clients[wp.Current]
  wp.Current = wp.Current + 1
  return client
}

func (wp *WorkerPool) UpdateHealth() {
    fmt.Println("Node: ",wp.Name, " retry active")
    conn := wp.GetConnection()

    conn.Mux.Lock()
    defer conn.Mux.Unlock()
    if _, err := conn.Client.NetworkID(context.Background()); err != nil {
      fmt.Println("Try connect to host: ",wp.Name, ", Error: ", err)
      return
    }

    //Active ok
    wp.mutex.Lock()
    defer wp.mutex.Unlock()
    wp.Active = true
    fmt.Println("Node: ",wp.Name, " succesfully reactive")

}
