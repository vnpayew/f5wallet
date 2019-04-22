import requests
import datetime

def RegisterAccount(host):
    url = "http://{}/api/v2/wallet/register".format(host)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Summary(host):
    url = "http://{}/api/v2/wallet/summary".format(host)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def ListWallet(host):
    url = "http://{}/api/v2/wallet/list".format(host)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)


############### Wallet functions ######################
def Create(host,name):
    typeWallet = 1
    url = "http://{}/api/v2/wallet/create/{}/{}".format(host,name,typeWallet)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Activate(host,name):
    state = 1
    url = "http://{}/api/v2/wallet/set_state/{}/{}".format(host,name,state)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Balance(host,name):
    url = "http://{}/api/v2/wallet/balance/{}".format(host,name)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Credit(host,name,amount):
    tx = datetime.datetime.now()
    url = "http://{}/api/v2/wallet/credit/{}/{}/{}".format(host,tx,name,amount)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Debit(host,name,amount):
    tx = datetime.datetime.now()
    url = "http://{}/api/v2/wallet/debit/{}/{}/{}".format(host,tx,name,amount)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)

def Transfer(host,sender,receiver, amount):
    tx = datetime.datetime.now()
    txType = 1
    url = "http://{}/api/v2/wallet/transfer/{}/{}/{}/".format(host,tx,sender,receiver,amount,txType)
    start = datetime.datetime.now()
    response = requests.get(url)
    diff = datetime.datetime.now() - start
    print(diff.total_seconds(), response.content)
