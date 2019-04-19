package subscribe

import (
  // "fmt"
  // "encoding/json"
  // "time"

  // "github.com/go-redis/redis"
  "sync"
  	// "log"
    // "runtime"
)

type BlockList struct {
    Blocks map[string]string
    mutex sync.RWMutex
}

func NewBlockList() *BlockList{
  blocks := make(map[string]string)
  return &BlockList{
    Blocks: blocks,
  }
}
func (bl *BlockList) Get(number string) (string, bool){
  bl.mutex.Lock()
  defer bl.mutex.Unlock()
  v,k := bl.Blocks[number]
  return string(v), k
}
func (bl *BlockList) Set(number string,value string){
  bl.mutex.Lock()
  defer bl.mutex.Unlock()
  bl.Blocks[number] = value
}
