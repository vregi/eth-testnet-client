package main

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	address := common.HexToAddress(os.Getenv("ADDRESS"))
	client, err := ethclient.Dial(os.Getenv("NET_URL"))
	if err != nil {
		log.Fatal(err)
	}

	checkAccountBalance(client, address)
}

func checkAccountBalance(client *ethclient.Client, address common.Address) {
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}

	ethValue := weiToEther(balance)

	log.Printf("Balance: %v ETH", ethValue)
}

func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}
