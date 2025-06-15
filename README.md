# simple ethereum client in go

a basic ethereum wallet cli app made for a university blockchain assignment. shows how to do basic wallet stuff like generating accounts, checking balances, signing messages, and sending transactions.

## what it does

- generates new ethereum accounts (private key, public key, address)
- checks account balances
- signs a message ("Hello, Ethereum!") and verifies the signature
- sends 0.01 ETH to another address

## how to use

1. create a `.env` file:
```
NET_URL=https://your-ethereum-rpc-url
PRIVATE_KEY=your-private-key-hex
TO_ADDRESS=0x1234567890123456789012345678901234567890
```

2. run it:
```bash
go run main.go
```

## described

- how ethereum addresses are generated from keys
- how message signing works with ethereum's standard
- how to build and send ethereum transactions
- basic cryptography with ecdsa keys

built with `go-ethereum` library for blockchain class.
