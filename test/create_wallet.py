from f5 import *

HOST_NAME="localhost:8080"
#url = "http://localhost:8080/api/v1/wallet/transfer/ffbcd481c1330e180879b4d2b9b50642eea43c02/a17a7a153c8d873a1df803c74e0664c13726f5e8/2/Test"

def create_wallets(host):
    list = "VI01,VI02,VI03"
    print("Start create wallets: ")
    names  = list.split(",")
    for name in names:
        print("Start wallet: ",name)
        Create(host,name)
    print("Waiting transactions 2 second")
    from time import sleep
    sleep(2) # Time in seconds.
    print("Start activate wallets: ",list)
    for name in names:
        print("Activate wallet: ",name)
        Activate(host,name)

    print("Waiting transactions 2 second")
    from time import sleep
    sleep(2) # Time in seconds.

    print("Start credit wallets: ",list, " 1000")
    for name in names:
        print("Credit wallet: ",name)
        Credit(host,name,1000)

def main():
    create_wallets(HOST_NAME)
    from time import sleep
    sleep(2) # Time in seconds.
    ListWallet(HOST_NAME)

if __name__== "__main__":
  main()
