package subscribe

import (
  "fmt"
  // "encoding/json"
  // "time"

  // "github.com/go-redis/redis"
  "github.com/ethereum/go-ethereum/core/types"
  "sync"
  	// "log"
    // "runtime"
)

type BlockList struct {
    Blocks map[string][]string
    //list  map[string]*types.Header
    mutex sync.RWMutex
}

func NewBlockList() *BlockList{
  blocks := make(map[string][]string)
  return &BlockList{
    Blocks: blocks,
  }
}

func (bl *BlockList) Get(number string) ([]string, bool){
  bl.mutex.Lock()
  defer bl.mutex.Unlock()
  values,k := bl.Blocks[number]
  return values, k
}
func (bl *BlockList) AddBlock(header *types.Header, value string) {
  bl.mutex.Lock()
  defer bl.mutex.Unlock()

  //Query redis
  blNumber := header.Number.String()

  if values, ok := bl.Blocks[blNumber]; ok {
       for _, v := range values {
          if (v == value) {
            fmt.Println("Value existed in list of block: ",blNumber)
            return
          }
       }
       fmt.Println("Add value:", value, " into list of block: ",blNumber)
       values = append(values, value)
       bl.Blocks[blNumber] = values
  } else {
      fmt.Println("Not find blockNumber:",blNumber,"Add first element: ", value)
      bl.Blocks[blNumber] = []string{value,}
  }
}
