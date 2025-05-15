package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
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

	privateKey, _, _ := account()
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	message := "Hello, Ethereum!"

	signature, err := signMessage(privateKey, message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Signature: %s", signature)

	valid, err := verifySignature(message, signature, publicKeyECDSA)
	if err != nil {
		log.Fatal(err)
	}

	if valid {
		log.Printf("Signature valid")
	} else {
		log.Printf("Signature invalid")
	}
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

func signMessage(privateKey *ecdsa.PrivateKey, message string) ([]byte, error) {
	signature, err := crypto.Sign(hashMessage(message), privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func hashMessage(message string) []byte {
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	messageHash := crypto.Keccak256([]byte(prefixedMessage))

	return messageHash
}

func verifySignature(message string, signature []byte, publicKey *ecdsa.PublicKey) (bool, error) {

	messageHash := hashMessage(message)
	sigPublicKey, err := crypto.SigToPub(messageHash, signature)
	if err != nil {
		return false, err
	}

	return publicKey.X.Cmp(sigPublicKey.X) == 0 && publicKey.Y.Cmp(sigPublicKey.Y) == 0, nil // comparing keys based on elliptic curve
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
