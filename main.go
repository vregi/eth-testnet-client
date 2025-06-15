package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
	"strconv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		command = flag.String("command", "", "Command to execute: create-account, balance, sign-message, send-tx, check-tx")
		address = flag.String("address", "", "Address to check balance or send to")
		amount  = flag.String("amount", "", "Amount in ETH to send")
		message = flag.String("message", "", "Message to sign")
		txHash  = flag.String("tx", "", "Transaction hash to check")
	)
	flag.Parse()

	client, err := ethclient.Dial(os.Getenv("NET_URL"))
	if err != nil {
		log.Fatal(err)
	}

	switch *command {
	case "create-account":
		createAccountCLI()
	case "balance":
		if *address == "" {
			log.Fatal("Address is required for balance command")
		}
		balanceCLI(client, *address)
	case "sign-message":
		if *message == "" {
			log.Fatal("Message is required for sign-message command")
		}
		signMessageCLI(*message)
	case "send-tx":
		if *address == "" || *amount == "" {
			log.Fatal("Address and amount are required for send-tx command")
		}
		sendTxCLI(client, *address, *amount)
	case "check-tx":
		if *txHash == "" {
			log.Fatal("Transaction hash is required for check-tx command")
		}
		checkTxCLI(client, *txHash)
	default:
		fmt.Println("Available commands:")
		fmt.Println("  -command=create-account                    Create new account")
		fmt.Println("  -command=balance -address=<addr>          Check balance")
		fmt.Println("  -command=sign-message -message=<msg>      Sign message")
		fmt.Println("  -command=send-tx -address=<addr> -amount=<eth>  Send transaction")
		fmt.Println("  -command=check-tx -tx=<hash>              Check transaction status")
	}
}

func createAccountCLI() {
	_, address, pkHex := account()
	fmt.Printf("New account created:\n")
	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Printf("Private Key: %s\n", pkHex)
	fmt.Printf("Save your private key securely!\n")
}

func balanceCLI(client *ethclient.Client, addressStr string) {
	address := common.HexToAddress(addressStr)
	fmt.Printf("Checking balance for address: %s\n", addressStr)
	checkAccountBalance(client, address)
}

func signMessageCLI(message string) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY not found in environment")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("Error parsing private key:", err)
	}

	signature, err := signMessage(privateKey, message)
	if err != nil {
		log.Fatal("Error signing message:", err)
	}

	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Signature: %s\n", hexutil.Encode(signature))

	// verify signature
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	valid, err := verifySignature(message, signature, publicKeyECDSA)
	if err != nil {
		log.Fatal("Error verifying signature:", err)
	}

	if valid {
		fmt.Printf("Signature verification: VALID\n")
	} else {
		fmt.Printf("Signature verification: INVALID\n")
	}
}

func sendTxCLI(client *ethclient.Client, toAddress, amountStr string) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY not found in environment")
	}

	// convert ETH to Wei
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Fatal("Invalid amount:", err)
	}

	weiAmount := new(big.Int)
	weiAmount.SetString(fmt.Sprintf("%.0f", amountFloat*1e18), 10)

	fmt.Printf("Sending %s ETH to %s\n", amountStr, toAddress)

	txHash, err := sendTransaction(client, privateKeyHex, toAddress, weiAmount)
	if err != nil {
		log.Fatal("Error sending transaction:", err)
	}

	fmt.Printf("Transaction sent successfully!\n")
	fmt.Printf("Transaction Hash: %s\n", txHash)
	fmt.Printf("Check transaction status with: -command=check-tx -tx=%s\n", txHash)
}

func checkTxCLI(client *ethclient.Client, txHashStr string) {
	txHash := common.HexToHash(txHashStr)

	fmt.Printf("Checking transaction: %s\n", txHashStr)

	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal("Transaction not found:", err)
	}

	if isPending {
		fmt.Printf("Status: PENDING\n")
	} else {
		fmt.Printf("Status: CONFIRMED\n")
	}

	fmt.Printf("From: %s\n", getSenderAddress(tx))
	if tx.To() != nil {
		fmt.Printf("To: %s\n", tx.To().Hex())
	}
	fmt.Printf("Value: %s ETH\n", weiToEther(tx.Value()))
	fmt.Printf("Gas Price: %s Gwei\n", weiToGwei(tx.GasPrice()))
	fmt.Printf("Gas Limit: %d\n", tx.Gas())

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err == nil {
		fmt.Printf("Block Number: %d\n", receipt.BlockNumber.Uint64())
		fmt.Printf("Gas Used: %d\n", receipt.GasUsed)
		if receipt.Status == 1 {
			fmt.Printf("Transaction Status: SUCCESS\n")
		} else {
			fmt.Printf("Transaction Status: FAILED\n")
		}
	} else {
		fmt.Printf("Receipt not yet available (transaction may be pending)\n")
	}
}

func getSenderAddress(tx *types.Transaction) string {
	chainID := big.NewInt(1)
	signer := types.NewEIP155Signer(chainID)

	from, err := types.Sender(signer, tx)
	if err != nil {
		return "Unknown"
	}
	return from.Hex()
}

func weiToGwei(wei *big.Int) *big.Float {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return gwei
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

func sendTransaction(client *ethclient.Client, privateKeyHex string, toAddress string, weiValue *big.Int) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("error parsing private key: %v", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("error getting nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting gas price: %v", err)
	}
	gasLimit := uint64(21000) // default gas

	toAddr := common.HexToAddress(toAddress)
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddr,
		Value:    weiValue,
	})

	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("error sending transaction: %v", err)
	}

	return signedTx.Hash().Hex(), nil
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
