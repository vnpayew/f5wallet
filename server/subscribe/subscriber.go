package subscribe

import (
  "fmt"
  // "encoding/json"
  // "time"
  	"context"
  	"github.com/ethereum/go-ethereum/ethclient"
  "github.com/ethereum/go-ethereum/core/types"
  // "github.com/go-redis/redis"
  	// "log"
    // "runtime"
)
type Subscriber struct {
   parent *SubscriberPool
   Name string
   SocketUrl string
   Active bool
}

func NewSubscriber(name string, socketUrl string)  *Subscriber {
     //Create transaction
     subscriber :=  &Subscriber{
       Name: name,
       SocketUrl: socketUrl,
       Active: true,

     }
     return subscriber
}
func (sb *Subscriber)  SetParent(parent *SubscriberPool) {
    sb.parent = parent
}

func (sb *Subscriber) ListenBlockEvent(){
		fmt.Println("Subscriber:", sb.Name ,"Listening from: ", sb.SocketUrl)
		websocket, err := ethclient.Dial("ws://" + sb.SocketUrl)
		if err != nil {
				fmt.Println("Cannot connect to websocket: ", err)
        sb.Active = false
        sb.parent.UpdatePolicyLevel()
				return
		}
		headers := make(chan *types.Header)
		sub, err := websocket.SubscribeNewHead(context.Background(), headers)
		if err != nil {
		    fmt.Println("Cannot SubscribeNewHead to host: ", sb.SocketUrl ," Error: ",err)
				return
		}
    sb.Active = true
	  fmt.Println("Start listening: ",sb.SocketUrl,"  ")
		for {
					select {
								case err := <-sub.Err():
										fmt.Println("Error from: ",sb.SocketUrl," Error: ",err)
                    sb.Active = false
										//log.Fatal(err)
								case header := <-headers:
                   fmt.Println("Block Number: ", header.Number.String()," Subscriber: ", sb.Name, " call CheckHeader")
                    //Process header
                    go func(){
                        sb.parent.UpdateBlockHeader(sb.Name, header)
                    }()
						}
		}
}
