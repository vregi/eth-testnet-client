package main

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

	account()
}

func account() (*ecdsa.PrivateKey, common.Address, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	pkBytes := crypto.FromECDSA(privateKey)
	log.Printf("Private key: %x", pkBytes)
	pkHex := hexutil.Encode(pkBytes)

	publicKey := privateKey.Public()
	pubECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	pubBytes := crypto.FromECDSAPub(pubECDSA)
	log.Printf("Public key: %x", pubBytes)

	address := crypto.PubkeyToAddress(*pubECDSA)

	return privateKey, address, pkHex
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
