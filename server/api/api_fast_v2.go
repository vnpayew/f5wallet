package api

import (
	"f5wallet/server/config"
	"f5wallet/server/handler"
	"f5wallet/server/rpc"
	"f5wallet/server/redis"
	"fmt"
	"github.com/qiangxue/fasthttp-routing"
  "strings"
	"math/big"
	// "github.com/ethereum/go-ethereum/common"
  // "github.com/go-redis/redis"
  // "encoding/json"
	  "strconv"
		"time"
		 "encoding/json"
)

type ApiFastV2 struct {
	 config * config.Config
	 walletHandler *handler.F5WalletHandler
	 redisHandler *handler.RedisHandler
}

func NewApiFastV2(cfg *config.Config, client *rpc.RpcRouting, rcache *redis.RedisPool) *ApiFastV2{
			whandler := handler.NewF5WalletHandler(cfg,rcache, cfg.F5Contract.Address, client)
			rhandler := handler.NewRedisHandler(rcache)
      return &ApiFastV2{
					config: cfg,
	        walletHandler:whandler,
					redisHandler:rhandler,
      }
}

// createTodo add a new todo
func (api *ApiFastV2) ProcessCall(c *routing.Context) error {
  method := c.Param("method")
  switch method {
			case "create":
           fmt.Println("call create")
           api.create(c)
           return  nil
			 case "list":
            fmt.Println("call list_wallet")
            api.list(c)
            return  nil
			case "history":
	           fmt.Println("call history")
	           api.history(c)
	           return  nil
       case "balance":
           fmt.Println("call balance")
           api.balance(c)
           return nil
			 case "state":
           fmt.Println("call state")
           api.state(c)
           return nil
       case "set_state":
           fmt.Println("call set_state")
           api.set_state(c)
           return nil
			 case "debit":
           fmt.Println("call debit")
           api.debit(c)
           return nil
			 case "credit":
           fmt.Println("call credit")
           api.credit(c)
           return nil
			 case "transfer":
 					fmt.Println("call transfer")
 					api.transfer(c)
 					return nil
			case "new_account":
           fmt.Println("call new_account")
           api.new_account(c)
           return nil
			 case "autofill":
		 			 fmt.Println("call autofill")
		 			 api.autofill(c)
		 			 return nil
			 case "register":
		 			 fmt.Println("call register")
		 			 api.registerAccounts(c)
		 			 return nil
			 case "summary":
					 fmt.Println("call summary ")
					 api.summary(c)
					 return nil
       case "accounts":
           fmt.Println("call accounts")
           api.accounts(c)
           return nil
			 case "test":
           fmt.Println("call test")
           fmt.Fprintf(c, "data=test")
           return nil
			 case "eth_transfer":
            fmt.Println("call eth_transfer")
            api.eth_transfer(c)
            return  nil
      case "eth_balance":
            fmt.Println("call eth_balance")
            api.eth_balance(c)
            return nil
   }

   fmt.Fprintf(c, "URL not found ")
   return nil
 }
 func (api *ApiFastV2) history(c *routing.Context){
	 hist_type := c.Param("p1")
	 records := []string{}
	 switch hist_type {
		 case "credit":
		   records  = api.walletHandler.CreditHistory()
		 case "debit":
			 records  = api.walletHandler.DebitHistory()
		 case "transfer":
 			records  = api.walletHandler.TransferHistory()
	 }
	 list_json, _ := json.Marshal(records)
	 fmt.Fprintf(c,string(list_json))
 }

 func (api *ApiFastV2) list(c *routing.Context){
	  ret_list := api.walletHandler.StashNames()
		list_json, _ := json.Marshal(ret_list)
		fmt.Fprintf(c,string(list_json))
 }

 func (api *ApiFastV2) summary(c *routing.Context){
	  n_account, n_wallet, n_credit, n_debit, n_transfer := api.walletHandler.GetSummary()
		ret_list := []string { "Number of registered Eth Accounts: " + strconv.Itoa(int(n_account)),
			" Number of wallet: " + n_wallet.String() ,
			" Number of credit transactions: " + n_credit.String() ,
			" Number of debit transactions: " +  n_debit.String() ,
			"  Number of transfer transactions: " +  n_transfer.String()}
		list_json, _ := json.Marshal(ret_list)
		fmt.Fprintf(c,string(list_json))
 }
func (api *ApiFastV2)  autofill(c *routing.Context){
			api.walletHandler.LoadAccountEth()
			ret_list := api.walletHandler.AutoFillGas()
			list_json, _ := json.Marshal(ret_list)
			fmt.Fprintf(c,string(list_json))
 }
func (api *ApiFastV2)  registerAccounts(c *routing.Context){
		 api.walletHandler.LoadAccountEth()
		 api.walletHandler.AutoFillGas()
		 requestTime := time.Now().UnixNano()
		 list := api.walletHandler.RegisterBatchEthToContract(requestTime)
		 list_json, _ := json.Marshal(list)
		 fmt.Fprintf(c,string(list_json))
}
 // call create wallet
 func (api *ApiFastV2)  create(c *routing.Context){
     account := c.Param("p1")
     account = strings.TrimPrefix(account,"0x")
		 typeStash := c.Param("p2")

		 typewallet, err :=  strconv.Atoi(typeStash)
		 if err != nil {
         fmt.Fprintf(c,"error: %v",err)
         return
     }
		 requestTime := time.Now().UnixNano()
     tx, err := api.walletHandler.CreateStash(requestTime,account,int8(typewallet))
     if err != nil {
         fmt.Fprintf(c,"error: %v",err)
         return
     }
     fmt.Fprintf(c,"transaction hash: ",tx.Hash().Hex())
 }

 // call balance of wallet
 func (api *ApiFastV2)  balance(c *routing.Context){
 		account := c.Param("p1")
 		account = strings.TrimPrefix(account,"0x")

 		bal, err := api.walletHandler.GetBalance(account)
 		if err != nil {
				fmt.Println("Error in call GetBalance:", err)
 				fmt.Fprintf(c,"error: ")
 				return
 		}
		fmt.Println("Return value:", bal)
 		fmt.Fprintf(c,"balance: %d",bal)
 }
 // call get wallet state
 func (api *ApiFastV2) state(c *routing.Context){
		account := c.Param("p1")
		account = strings.TrimPrefix(account,"0x")

		state, err := api.walletHandler.GetState(account)
		if err != nil {
			  fmt.Println("Error in state: ",err)
				fmt.Fprintf(c,"error:")
				return
		}
		fmt.Fprintf(c,"transaction hash: ",state)
 }
 // call set wallet state
 func (api *ApiFastV2) set_state(c *routing.Context){
		 account := c.Param("p1")
		 account = strings.TrimPrefix(account,"0x")
		 state := c.Param("p2")

		 stashState, err := strconv.Atoi(state)
		 if err != nil {
			fmt.Fprintf(c,"error: Please txType as integer ")
			return
		}
		requestTime := time.Now().UnixNano()
		 tx, err := api.walletHandler.SetState(requestTime,account,int8(stashState))
		 if err != nil {
				 fmt.Fprintf(c,"error: %v",err)
				 return
		 }
		 fmt.Fprintf(c,"transaction hash: ",tx.Hash().Hex())
 }

 // call get wallet state
 func (api *ApiFastV2) debit(c *routing.Context){
		txRef := c.Param("p1")
		account := c.Param("p2")
		account = strings.TrimPrefix(account,"0x")
		value := c.Param("p3")

		amount := new(big.Int)
		amount.SetString(value,10)

		requestTime := time.Now().UnixNano()
		tx, err := api.walletHandler.Debit(requestTime,txRef,account,amount)
		if err != nil {
				fmt.Fprintf(c,"error: %v",err)
				return
		}
		fmt.Fprintf(c,"transaction hash: ",tx.Hash().Hex())
 }
 // call get wallet state
 func (api *ApiFastV2) credit(c *routing.Context){
	 txRef := c.Param("p1")
	 account := c.Param("p2")
	 account = strings.TrimPrefix(account,"0x")
	 value := c.Param("p3")

	 amount := new(big.Int)
	 amount.SetString(value,10)

	 requestTime := time.Now().UnixNano()
	 tx, err := api.walletHandler.Credit(requestTime,txRef,account,amount)
	 if err != nil {
			 fmt.Fprintf(c,"error: %v",err)
			 return
	 }
	 fmt.Fprintf(c,"transaction hash: ",tx.Hash().Hex())
 }
 // call transfer token
 func (api *ApiFastV2) transfer(c *routing.Context){
	   txRef := c.Param("p1")
     sender := c.Param("p2")
     receiver := c.Param("p3")
     value := c.Param("p4")
     note := c.Param("p5")
		 txtyp := c.Param("p6")

     if sender == "" {
       fmt.Fprintf(c,"error: Please add sender address ")
       return
     }
     if receiver == "" {
       fmt.Fprintf(c,"error: Please add receiver address ")
       return
     }
		 amount := new(big.Int)
		 amount.SetString(value,10)

		 txType, err :=  strconv.Atoi(txtyp)
		 if err != nil {
			 fmt.Fprintf(c,"error: Please txType as integer ")
			 return
		 }

		 requestTime := time.Now().UnixNano()
  	 result, err := api.walletHandler.Transfer(requestTime, txRef,sender,receiver,amount,note,int8(txType))
     if err != nil {
           fmt.Fprintf(c,"Error to transfer token: %v", err)
           return
     }
		 fmt.Fprintf(c,"transaction: %v ", result.Hash().Hex())
     // fmt.Fprintf(c,"transaction: penđing")
 }
 // call transfer token
 // func (api *ApiFastV2) report(c *routing.Context){
 //     fmt.Println("Start report")
 //     report := api.redisHandler.Report()
 //
 //     fmt.Fprintf(c,"data:" + report)
 // }
 func (api *ApiFastV2) new_account(c *routing.Context){
     account, err := api.walletHandler.NewAccountEth()
     if err != nil {
       // handle error
			 fmt.Fprintf(c,"error: %v",err )
			 return
     }
    fmt.Fprintf(c,"account: %v",account )
 }

 func (api *ApiFastV2) accounts(c *routing.Context){
     accounts := api.walletHandler.GetAccountList()
  	 list := []string{}
		 for _,account := range accounts {
			   addr := account.Hex()
				 addr = strings.ToLower(strings.TrimPrefix(addr,"0x"))
				 list = append(list,addr)
		 }
    fmt.Fprintf(c,"accounts: %v",list )
 }
  // call transfer eth
  func (api *ApiFastV2)  eth_transfer(c *routing.Context){
      from := c.Param("p1")
      to := c.Param("p2")
      amount := c.Param("p3")

      if from == "" {
        fmt.Fprintf(c,"error: Please add from address ")
        return
      }
      if to == "" {
        fmt.Fprintf(c,"error: Please add to address ")
        return
      }
      from = strings.TrimPrefix(from,"0x")
      to = strings.TrimPrefix(to,"0x")

   	 result, err := api.walletHandler.EthTransfer(from,to,amount)
      if err != nil {
            fmt.Fprintf(c,"Error to transfer token: %v", err)
            return
      }
 		 fmt.Fprintf(c,"transaction: %v ", result)
      // fmt.Fprintf(c,"transaction: penđing")
  }

  // call transfer token
  func (api *ApiFastV2) eth_balance(c *routing.Context){
      account := c.Param("p1")
      account = strings.TrimPrefix(account,"0x")

      bal, err := api.walletHandler.EthBalaneOf(account)
      if err != nil {
          fmt.Fprintf(c,"error: %v",err)
          return
      }
      fmt.Fprintf(c,"balance: %d",bal)
  }
