package handler

import (
  "f5wallet/f5coin"
  "f5wallet/server/rpc"
  "f5wallet/server/account"
  "f5wallet/server/config"
  "f5wallet/server/redis"
  // "time"
  "math/big"
  // "strings"
  "fmt"
  // "encoding/json"
  "errors"
  "strings"
  "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  "github.com/ethereum/go-ethereum/core/types"
  "github.com/ethereum/go-ethereum/accounts/abi/bind"
  "github.com/jinzhu/gorm"
  "encoding/hex"
  "sync"
  "math"
  "strconv"
  "bytes"
)

type F5WalletHandler struct {
    cfg *config.Config
    redisCache *redis.RedisPool
    Client *rpc.RpcRouting
    Wallets []*account.WalletAccount
    ContractAddress common.Address
    Current int
    Mutex sync.Mutex
}


//// ####################################### Support function ################
func stringTo32Byte(data string) [32]byte {
  //hexstring := hex.EncodeToString([]byte(data))
  var arr [32]byte
	copy(arr[:], data)
  return arr
}

func byte32ToString(data [32]byte) string {
  return string(bytes.Trim(data[:], "\x00"))
}

func isConnectionError(err error) bool {
    err_msg := err.Error()
    if strings.Contains(err_msg, "connection refused") {
        return true
    }
    return false
}

//// ####################################### Processing Support function ################
func NewF5WalletHandler(cfg *config.Config, rc *redis.RedisPool, contract_address string, client *rpc.RpcRouting)  *F5WalletHandler{
      contractAddress := common.HexToAddress(contract_address)
      wallHandler :=  &F5WalletHandler{
        cfg: cfg,
        redisCache: rc,
        Client: client,
        ContractAddress: contractAddress,
        Current: 0,
      }
      wallHandler.LoadAccountEth()
      wallHandler.AutoFillGas()
      // wallHandler.RegisterBatchEthToContract()
      return wallHandler
}

func (fw *F5WalletHandler) RegisterBatchEthToContract(requestTime int64) []string {
    ret := []string{}
    list := fw.GetAccountList()
    sublist :=  []common.Address{}
    for i,item := range list {
      sublist = append(sublist,item)
      if i > 0 && i % 5 == 0 {
        if len(sublist) > 0 {
          fmt.Println("F5WalletHandler.RegisterBatchEthToContract: Start register sublist: ", i/5)
          tx,err := fw.RegisterAccETH(requestTime,sublist)
          if err != nil {
             ret = append(ret, err.Error())
          } 	else {
             ret = append(ret, tx.Hash().Hex())
          }
          sublist = []common.Address{}
        }
      }
    }
    if len(sublist) > 0 {
      fmt.Println("F5WalletHandler.RegisterBatchEthToContract: Start register last sublist")
      tx,err := fw.RegisterAccETH(requestTime,sublist)
      if err != nil {
         ret = append(ret, err.Error())
      } 	else {
         ret = append(ret, tx.Hash().Hex())
      }
    }
    return ret
}


//// ####################################### Blockchain call function ################
func  (fw *F5WalletHandler) CreditHistory() []string {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    ret, err := fw.creditHistory(conn)
    if err == nil {
       return ret
    }
    if !isConnectionError(err) {
        return ret
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call creditHistory 3 times. ")
  return []string{}
}

func (fw *F5WalletHandler)  creditHistory(conn *rpc.RpcConnection) ([]string, error) {
  ret := []string{}
  instance, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)

  n_credit, err := instance.GetCreditHistoryLength(&bind.CallOpts{From: owner})
  if err != nil {
    fmt.Println("Cannot Get length of wallets, error: ",err)
    return ret,err
  }
  i := int64(0)
  for i < n_credit.Int64() {
      creditId, err := instance.CreditIdx(&bind.CallOpts{From: owner},big.NewInt(i))
      i = i + 1
      if(err != nil) {
         fmt.Println("Error get Debit Idx: ", err)
         continue
      }
      creditTx, err := instance.Credits(&bind.CallOpts{From: owner},creditId)
      if(err != nil) {
         fmt.Println("Error get Debit Transactions: ", string(creditId[:])," Error: ", err)
         continue
      }
      list := []string{
        byte32ToString(creditTx.TxRef),
        byte32ToString(creditTx.StashName),
        creditTx.Amount.String(),
        creditTx.Timestamp.String(),
      }
      credit_tx :=  strings.Join(list, ",")
      ret = append(ret,credit_tx)
  }
  return ret, nil
}

func  (fw *F5WalletHandler) DebitHistory() []string {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    ret, err := fw.debitHistory(conn)
    if err == nil {
       return ret
    }
    if !isConnectionError(err) {
        return ret
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call debitHistory 3 times. ")
  return []string{}
}
func  (fw *F5WalletHandler) debitHistory(conn *rpc.RpcConnection) ([]string,error) {
  ret := []string{}
  instance, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)

  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
  n_debit, err := instance.GetDebitHistoryLength(&bind.CallOpts{From: owner})
  if err != nil {
    fmt.Println("Cannot Get length of wallets, error: ",err)
    return ret,err
  }
  i := int64(0)
  for i < n_debit.Int64() {
      debitId, err := instance.DebitIdx(&bind.CallOpts{From: owner},big.NewInt(i))
      i = i + 1
      if(err != nil) {
         fmt.Println("Error get Debit Idx: ", err)
         continue
      }
      debitTx, err := instance.Debits(&bind.CallOpts{From: owner},debitId)
      if(err != nil) {
         fmt.Println("Error get Debit Transactions: ", string(debitId[:])," Error: ", err)
         continue
      }
      list := []string{
        byte32ToString(debitTx.TxRef),
        byte32ToString(debitTx.StashName),
        debitTx.Amount.String(),
        debitTx.Timestamp.String(),
      }
      debit_tx :=  strings.Join(list, ",")
      ret = append(ret,debit_tx)
  }
  return ret,nil
}
func  (fw *F5WalletHandler) TransferHistory() []string {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    ret, err := fw.transferHistory(conn)
    if err == nil {
       return ret
    }
    if !isConnectionError(err) {
        return ret
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return []string{}
}
func  (fw *F5WalletHandler) transferHistory(conn *rpc.RpcConnection) ([]string,error) {
  ret := []string{}

  instance, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)

  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
  n_transfer, err := instance.GetTransferHistoryLength(&bind.CallOpts{From: owner})
  if err != nil {
    fmt.Println("Cannot Get length of wallets, error: ",err)
    return ret, err
  }
  i := int64(0)
  for i < n_transfer.Int64() {
      transferId, err := instance.TransferIdx(&bind.CallOpts{From: owner},big.NewInt(i))
      i = i + 1
      if(err != nil) {
         fmt.Println("Error get Debit Idx: ", err)
         continue
      }
      transferTx, err := instance.Transfers(&bind.CallOpts{From: owner},transferId)
      if(err != nil) {
         fmt.Println("Error get Debit Transactions: ", string(transferId[:])," Error: ", err)
         continue
      }
      list := []string{
        byte32ToString(transferTx.TxRef),
        byte32ToString(transferTx.Sender),
        byte32ToString(transferTx.Receiver),
        transferTx.Amount.String(),
        transferTx.Note,
        strconv.Itoa(int(transferTx.TxType)),
        transferTx.Timestamp.String(),
      }
      transfer_tx :=  strings.Join(list, ",")
      ret = append(ret,transfer_tx)
  }
  return ret,nil
}

func  (fw *F5WalletHandler) StashNames() []string {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    ret, err := fw.stashNames(conn)
    if err == nil {
       return ret
    }
    if !isConnectionError(err) {
        return ret
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return []string{}
}

func  (fw *F5WalletHandler) stashNames(conn *rpc.RpcConnection) ([]string,error) {
  ret := []string{}

  instance, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)

  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
  n_wallet, err := instance.GetStashNamesLenght(&bind.CallOpts{From: owner})
  if err != nil {
    fmt.Println("Cannot Get length of wallets, error: ",err)
    return ret,err
  }
  i := int64(0)
  for i < n_wallet.Int64() {
      stash_name, err := instance.StashNames(&bind.CallOpts{From: owner},big.NewInt(i))
      i = i + 1
      if(err != nil) {
         fmt.Println("Error get StashNames: ", err)
         continue
      }
      bal, err := instance.GetBalance(&bind.CallOpts{From: owner},stash_name)
      if(err != nil) {
         fmt.Println("Error get balance of: ", string(stash_name[:]), "Error:", err)
         continue
      }
      state, err := instance.GetState(&bind.CallOpts{From: owner},stash_name)
      if(err != nil) {
         fmt.Println("Error get state of: ", string(stash_name[:]), "Error:", err)
         continue
      }
      list := []string{
         byte32ToString(stash_name),
         bal.String(),
         strconv.Itoa(int(state)),
      }
      wallet_info := strings.Join(list,",")
      ret = append(ret,wallet_info)
  }
  return ret,nil
}


func  (fw *F5WalletHandler)  GetSummary() (int16,*big.Int, *big.Int, *big.Int,*big.Int)   {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    n_account, n_wallet, n_credit, n_debit, n_transfer, err := fw.getSummary(conn)
    if err == nil {
       return n_account, n_wallet, n_credit, n_debit, n_transfer
    }
    if !isConnectionError(err) {
        return n_account, n_wallet, n_credit, n_debit, n_transfer
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return 0,nil,nil,nil,nil
}

func  (fw *F5WalletHandler) getSummary(conn *rpc.RpcConnection) (int16,*big.Int, *big.Int, *big.Int,*big.Int, error)   {
      instance, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)

      owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
      n_account, err := instance.GetRegistedAccEthLength(&bind.CallOpts{From: owner})

      if err != nil {
        fmt.Println("Cannot Get Registed Acc Eth Length error: ",err, ", Contract: ", fw.ContractAddress.Hex())

        return 0, nil, nil, nil, nil, err
      }
      n_wallet, err := instance.GetStashNamesLenght(&bind.CallOpts{From: owner})
      if err != nil {
        fmt.Println("Cannot Get length of wallets, error: ",err)
        return n_account, nil, nil, nil, nil,err
      }
      n_credit, err := instance.GetCreditHistoryLength(&bind.CallOpts{From: owner})
      if err != nil {
        fmt.Println("Cannot Get length of credits, error: ",err)
        return n_account, n_wallet, nil, nil, nil,err
      }
      n_debit, err := instance.GetDebitHistoryLength(&bind.CallOpts{From: owner})
      if err != nil {
        fmt.Println("Cannot Get length of debit error: ",err)
        return n_account, n_wallet, n_credit, nil, nil,err
      }
      n_transfer, err := instance.GetTransferHistoryLength(&bind.CallOpts{From: owner})
      if err != nil {
        fmt.Println("Cannot Get length of transfer, error: ",err)
        return n_account, n_wallet, n_credit, n_debit, nil,err
      }
      return  n_account, n_wallet, n_credit, n_debit, n_transfer,nil
}
func (fw *F5WalletHandler) GetBalance(stashName string) (*big.Int, error)  {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    bal, err := fw.getBalance(conn,stashName)
    if err == nil {
       return bal,nil
    }
    if !isConnectionError(err) {
        return bal,err
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return nil,errors.New("Connection errors")
}
func (fw *F5WalletHandler) getBalance(conn *rpc.RpcConnection, stashName string) (*big.Int, error)  {
    fmt.Println("F5WalletHandler.GetBalance: Start get balance ")
    conn.Mux.Lock()
    defer  conn.Mux.Unlock()
    session,err  := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
    if err != nil {
        fmt.Println("Cannot find F5 contract")
        return nil,err
    }
    fmt.Println("F5WalletHandler.GetBalance: call  GetBalance")

    owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)

    return session.GetBalance(&bind.CallOpts{From: owner},stringTo32Byte(stashName))
}

func (fw *F5WalletHandler) GetState(stashName string) (int8, error)  {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    bal, err := fw.getState(conn,stashName)
    if err == nil {
       return bal,nil
    }
    if !isConnectionError(err) {
        return bal,err
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return 0,errors.New("Connection errors")
}
func (fw *F5WalletHandler) getState(conn *rpc.RpcConnection, stashName string) (int8, error)  {
  conn.Mux.Lock()
  defer  conn.Mux.Unlock()

  session, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
  if err != nil {
      fmt.Println("Cannot find F5 contract")
      return 0,err
  }
  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
  return session.GetState(&bind.CallOpts{From: owner},stringTo32Byte(stashName))
}

func (fw *F5WalletHandler) GetTransferHistoryLength() (*big.Int, error)  {
  retry := 0
  for retry < 3 {
    conn := fw.Client.GetConnection()
    n, err := fw.getTransferHistoryLength(conn)
    if err == nil {
       return n,nil
    }
    if !isConnectionError(err) {
        return n,err
    }
    fw.Client.DeactiveNode(conn.Name)
    retry = retry + 1
  }
  fmt.Println("Failed to retry call transferHistory 3 times. ")
  return big.NewInt(0),errors.New("Connection errors")
}
func (fw *F5WalletHandler) getTransferHistoryLength(conn *rpc.RpcConnection) (*big.Int, error)  {
  conn.Mux.Lock()
  defer  conn.Mux.Unlock()
  session,err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
  if err != nil {
      fmt.Println("Cannot find F5 contract")
      return nil,err
  }
  owner := common.HexToAddress("0x"+ fw.cfg.F5Contract.Owner)
  return session.GetTransferHistoryLength(&bind.CallOpts{From:owner})
}

////////////////////////////////////////////// Transaction function ##################//////////////////////////////////

func (fw *F5WalletHandler) RegisterAccETH(requestTime int64, listAcc []common.Address) (*types.Transaction, error) {
    fmt.Println("F5WalletHandler.RegisterAccETH: Start RegisterAccETH, account: ", fw.cfg.F5Contract.Owner)
    account := fw.GetAccountEthAddress(fw.cfg.F5Contract.Owner)
    if account == nil {
       fmt.Println("F5WalletHandler.RegisterAccETH: Cannot find active account")
       return nil, errors.New("Cannot find bugdet account")
    }
    fmt.Println("F5WalletHandler.RegisterAccETH: Get new transactor")
    auth := account.NewTransactor()

    fmt.Println("F5WalletHandler.RegisterAccETH: Get gas limit")
    gasLimit,ok := fw.cfg.F5Contract.GasLimit["register"]
    if !ok {
      gasLimit = fw.cfg.F5Contract.GasLimitDefault
    }
    auth.GasLimit = gasLimit

    retry := 0
    for retry < 3 {
      fmt.Println("F5WalletHandler.RegisterAccETH: Start RegisterAccETH, retry: ", retry)
      conn := fw.Client.GetConnectionByAccount(account.Address)

      fmt.Println("F5WalletHandler.RegisterAccETH: call registerAccETH ", len(listAcc))
      tx, err := fw.registerAccETH(conn,auth,listAcc)
      if err == nil {
         key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
         fw.redisCache.LogStart(key, 0, requestTime )
         return tx,nil
      }
      if !isConnectionError(err) {
          return tx,err
      }
      fmt.Println("F5WalletHandler.RegisterAccETH: deactive node: ", conn.Name)
      fw.Client.DeactiveNode(conn.Name)
      retry = retry + 1
    }
    fmt.Println("F5WalletHandler.RegisterAccETH: End RegisterAccETH: retry failed ")
    return nil, errors.New("Cannot find wallet in pool to create transaction")
}
func (fw *F5WalletHandler) registerAccETH(conn *rpc.RpcConnection, auth *bind.TransactOpts, listAcc []common.Address) (*types.Transaction, error) {
      session,err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
      if err != nil {
          fmt.Println("Cannot find F5 contract")
          return nil,err
      }
      conn.Mux.Lock()
      defer  conn.Mux.Unlock()

      return session.RegisterAccETH(auth,listAcc)
}
func (fw *F5WalletHandler) CreateStash(requestTime int64, stashName string, typeStash int8) (*types.Transaction, error)  {
    retry := 0
    for retry <10 {
        account := fw.GetAccountEth()
        if account.IsAvailable() {
          conn := fw.Client.GetConnectionByAccount(account.Address)
          session, err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
          if err != nil {
              fmt.Println("Cannot find F5 contract")
              return nil,err
          }
          auth := account.NewTransactor()
          gasLimit,ok := fw.cfg.F5Contract.GasLimit["create"]
          if !ok {
            gasLimit = fw.cfg.F5Contract.GasLimitDefault
          }
          auth.GasLimit = gasLimit

          conn.Mux.Lock()
          defer  conn.Mux.Unlock()
          bs := stringTo32Byte(stashName)
          fmt.Println("Using: ", account, " to create Wallet: ",stashName, " len: ", len(bs) )
          tx,err := session.CreateStash(auth,bs, typeStash)
          if(err == nil){
            //Log transaction
            key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
            fw.redisCache.LogStart(key, 0, requestTime )
            return tx, err
          }
          if isConnectionError(err) {
              fw.Client.DeactiveNode(conn.Name)
          }
        }
        retry = retry + 1
    }
    return nil, errors.New("Cannot find wallet in pool to create transaction")
}
func (fw *F5WalletHandler) SetState(requestTime int64, stashName string, stashState int8 ) (*types.Transaction, error)  {
  retry := 0
  for retry <10 {
      account := fw.GetAccountEth()
      if account.IsAvailable() {
          auth := account.NewTransactor()
          gasLimit,ok := fw.cfg.F5Contract.GasLimit["state"]
          if !ok {
            gasLimit = fw.cfg.F5Contract.GasLimitDefault
          }
          auth.GasLimit = gasLimit
          conn := fw.Client.GetConnectionByAccount(account.Address)
          session, err  := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
          if err != nil {
              fmt.Println("Cannot find F5 contract")
              return nil,err
          }
          conn.Mux.Lock()
          defer  conn.Mux.Unlock()
          tx,err := session.SetState(auth, stringTo32Byte(stashName),stashState)
          if(err == nil){
            //Log transaction
            key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
            fw.redisCache.LogStart(key, 0, requestTime )
            return tx, err
          }
          if isConnectionError(err) {
              fw.Client.DeactiveNode(conn.Name)
          }
      }
        retry = retry + 1
  }
  return nil, errors.New("Cannot find wallet in pool to create transaction")
}

func (fw *F5WalletHandler) Debit(requestTime int64, txRef string, stashName string, amount *big.Int) (*types.Transaction, error) {
    retry := 0
    for retry <10 {
        account := fw.GetAccountEth()
        if account.IsAvailable() {
            auth := account.NewTransactor()
            gasLimit,ok := fw.cfg.F5Contract.GasLimit["debit"]
            if !ok {
              gasLimit = fw.cfg.F5Contract.GasLimitDefault
            }
            auth.GasLimit = gasLimit
            conn := fw.Client.GetConnectionByAccount(account.Address)
            session,err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
            if err != nil {
                fmt.Println("Cannot find F5 contract")
                return nil,err
            }
            conn.Mux.Lock()
            defer  conn.Mux.Unlock()
            tx,err := session.Debit(auth, stringTo32Byte(txRef),stringTo32Byte(stashName),amount)
            if(err == nil){
              //Log transaction
              key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
              fw.redisCache.LogStart(key, 0, requestTime )
              return tx, err
            }
            if isConnectionError(err) {
                fw.Client.DeactiveNode(conn.Name)
            }
        }
          retry = retry + 1
    }
    return nil, errors.New("Cannot find wallet in pool to create transaction")
}

func (fw *F5WalletHandler) Credit(requestTime int64, txRef string, stashName string, amount *big.Int) (*types.Transaction, error) {
  retry := 0
  for retry <10 {
      account := fw.GetAccountEth()
      if account.IsAvailable() {
          auth := account.NewTransactor()
          gasLimit,ok := fw.cfg.F5Contract.GasLimit["credit"]
          if !ok {
            gasLimit = fw.cfg.F5Contract.GasLimitDefault
          }
          auth.GasLimit = gasLimit
          conn := fw.Client.GetConnectionByAccount(account.Address)
          session,err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)

          if err != nil {
              fmt.Println("Cannot find F5 contract")
              return nil,err
          }
          conn.Mux.Lock()
          defer  conn.Mux.Unlock()
          tx,err :=  session.Credit(auth, stringTo32Byte(txRef),stringTo32Byte(stashName),amount)
          if(err == nil){
            //Log transaction
            key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
            fw.redisCache.LogStart(key, 0, requestTime )
            return tx, err
          }
          if isConnectionError(err) {
              fw.Client.DeactiveNode(conn.Name)
          }
      }
  }
  return nil, errors.New("Cannot find wallet in pool to create transaction")
}

func (fw *F5WalletHandler) Transfer(requestTime int64, txRef string, sender string, receiver string, amount *big.Int, note string, txType int8) (*types.Transaction, error) {
  retry := 0
  for retry <10 {
      account := fw.GetAccountEth()
      if account.IsAvailable() {
          auth := account.NewTransactor()
          gasLimit,ok := fw.cfg.F5Contract.GasLimit["transfer"]
          if !ok {
            gasLimit = fw.cfg.F5Contract.GasLimitDefault
          }
          auth.GasLimit = gasLimit
          conn := fw.Client.GetConnectionByAccount(account.Address)
          session,err := f5coin.NewBusiness(fw.ContractAddress,conn.Client)
          if err != nil {
              fmt.Println("Cannot find F5 contract")
              return nil,err
          }
          conn.Mux.Lock()
          defer  conn.Mux.Unlock()
          tx, err :=  session.Transfer(auth, stringTo32Byte(txRef),stringTo32Byte(sender),stringTo32Byte(receiver),amount,note,txType)
          if(err == nil){
            //Log transaction
            key := strings.TrimPrefix(tx.Hash().Hex(),"0x")
            fw.redisCache.LogStart(key, 0, requestTime )
            return tx, err
          }
          if isConnectionError(err) {
              fw.Client.DeactiveNode(conn.Name)
          }

      }
        retry = retry + 1
  }
  return nil, errors.New("Cannot find wallet in pool to create transaction")
}
//########################### Ethererum Coin Function ############################
func (fw *F5WalletHandler) EthBalaneOf(account string) (*big.Float,error) {
  wallet := fw.GetAccountEthAddress(account)
  if wallet != nil {
      return wallet.EthBalaneOf()
  }
  return nil, errors.New("Cannot find account in system")
}

func (fw *F5WalletHandler) EthTransfer(from string,to string,amount string) (string,error) {
   wallet := fw.GetAccountEthAddress(from)

   fromAddress := common.HexToAddress("0x" + wallet.Address)
   nonce, err := wallet.Routing.PendingNonceAt(fromAddress)
   if err != nil {
     fmt.Println("Error in getting nonce ")
     return "", err
   }

   gLimit := fw.cfg.Contract.GasLimit
   gPrice := fw.cfg.Contract.GasPrice

   gasLimit := uint64(gLimit)
   gasPrice := new(big.Int)
   gasPrice, _ = gasPrice.SetString(gPrice, 10)

   toAddress := common.HexToAddress("0x" + to)

   eth_unit := big.NewFloat(math.Pow10(18))
   amount_value := new(big.Float)
   value, ok := amount_value.SetString(amount)

   if !ok {
        fmt.Println("SetString: error")
        return "", errors.New("convert amount error")
   }
   value = value.Mul(value,eth_unit)

   value_transfer := new(big.Int)
   value.Int(value_transfer)

   var data []byte
   rawTx := types.NewTransaction(nonce, toAddress, value_transfer, gasLimit, gasPrice, data)

   signer := types.FrontierSigner{}
   signature, err := crypto.Sign(signer.Hash(rawTx).Bytes(), wallet.PrivateKey)
   if err != nil {
     fmt.Println(" Cannot sign contract: ", err)
     return "",err
   }

   signedTx, err := rawTx.WithSignature(signer, signature)

   txhash := strings.TrimPrefix(signedTx.Hash().Hex(),"0x")
   err = wallet.Routing.SubmitTransaction(signedTx,nonce)

   return txhash, err
}

func (fw *F5WalletHandler) AutoFillGas() []string {
    fw.Mutex.Lock()
    defer fw.Mutex.Unlock()

    ret := []string{}
    for _, wallet := range fw.Wallets {
      bal, err := wallet.EthBalaneOf()
      if err != nil {
         fmt.Println("Cannot get wallet balance. Deactive wallet")
         wallet.Active = false
         continue
      }
      ba,_ := bal.Float64()
      if ba < 1000 {
         fmt.Println("Create transaction to fillGass from budget")
         var fill_account int = int(1000 - ba)

         txhash, err := fw.EthTransfer(fw.cfg.F5Contract.EthBudget, wallet.Address,strconv.Itoa(fill_account))
         if err != nil {
           fmt.Println("Cannot fill more gas. Deactive wallet ")
           wallet.Active = false
           continue
         }
         fmt.Println("Fill Eth to account: ", wallet.Address, " transaction: ", txhash)
         ret = append(ret,txhash)
      } else {
         fmt.Println("Account: ", wallet.Address, " balance: ", ba)
      }
    }
    return ret
}

//#########################   Non blockchain function ##########################
func (fw *F5WalletHandler) NewAccountEth() (string, error) {
      privateKey, err := crypto.GenerateKey()
      if err != nil {
        return "",err
      }
      address := crypto.PubkeyToAddress(privateKey.PublicKey)

      acc := address.Hex()
      acc = strings.TrimPrefix(acc,"0x")
      acc = strings.ToLower(acc)

      priKey :=  hex.EncodeToString(crypto.FromECDSA(privateKey))

      new_account := &config.TokenAccount{
        Address: acc,
        PrivateKey: priKey,
        Active: true,
      }

      fmt.Println("Update account to db ")
      db, err := gorm.Open("mysql", fw.cfg.MysqlConnectionUrl())
      if fw.cfg.Mysql.Debug {
         db.LogMode(true)
      }

      if err != nil {
        panic("failed to connect database: " + err.Error())
      }
      defer db.Close()
      //fmt.Println("Create new record")
      db.Create(new_account)

      fmt.Println("Update account to wallet ")
      wallet := account.NewWalletAccount(fw.cfg, fw.Client, acc, 0, privateKey, new_account)

      fw.Wallets = append(fw.Wallets,wallet)
      return acc, nil
}

func (fw *F5WalletHandler) GetAccountEthAddress(addr string) *account.WalletAccount {
    for _, wallet := range fw.Wallets {
       if wallet.Address == addr {
         return wallet
       }
    }
    return nil
}

func (fw *F5WalletHandler) GetAccountEth() *account.WalletAccount{
    fw.Mutex.Lock()
    defer fw.Mutex.Unlock()
    len := len(fw.Wallets)
    if len == 0 {
      return nil
    }
    if fw.Current >= len {
         fw.Current = fw.Current % len
    }
    wallet := fw.Wallets[fw.Current]
    fw.Current = fw.Current + 1
    return wallet
}

func (fw *F5WalletHandler) LoadAccountEth(){
  fmt.Println("Start load accounts from db to create wallets ")
  db, err := gorm.Open("mysql", fw.cfg.MysqlConnectionUrl())
  if fw.cfg.Mysql.Debug {
     db.LogMode(true)
  }

  if err != nil {
    panic("failed to connect database: " + err.Error())
  }
  defer db.Close()

  accounts := []config.TokenAccount{}

  if err := db.Where("active = ?", true).Find(&accounts).Error; err != nil {
    fmt.Println("Cannot get active Token Account in db: ",err)
    return
  }

  wallets := []*account.WalletAccount{}
  for _, acc := range accounts {
      fmt.Println("Load wallet: ",acc.Address)
      b, err := hex.DecodeString(acc.PrivateKey)
     if err != nil {
          fmt.Println("invalid hex string: " + acc.PrivateKey)
         continue
     }
      privateKey := crypto.ToECDSAUnsafe(b)
      wallet := account.NewWalletAccount(fw.cfg, fw.Client, acc.Address, 0, privateKey, &acc)

      if fw.cfg.Webserver.NonceMode == 2 {
          fmt.Println("Start sync nonce of ",acc.Address)
          wallet.SyncNonce()
      }
      wallets = append(wallets,wallet)
  }
  fmt.Println("End load accounts from db: ", len(wallets))
  fw.Mutex.Lock()
  defer fw.Mutex.Unlock()
  fw.Wallets = wallets
}

func (fw *F5WalletHandler) GetAccountList() ([]common.Address) {
   fmt.Println("F5WalletHandler.GetAccountList: start read wallets")
   fw.Mutex.Lock()
   defer fw.Mutex.Unlock()
   accounts := []common.Address{}
   for _,wallet := range fw.Wallets {
       if wallet.Active {
         address := common.HexToAddress("0x"+wallet.Address)
         accounts = append(accounts,address)
       }
   }
   fmt.Println("F5WalletHandler.GetAccountList: end read wallets:",len(accounts))
   return accounts
}
