package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}

}

func run() error {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey).String()
	fmt.Println(address)

	v := struct {
		Name string
	}{
		Name: "Quincy",
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	txHash := crypto.Keccak256(data)

	// Sign the hash with the private key to produce a signature.
	sig, err := crypto.Sign(txHash, privateKey)
	if err != nil {
		return fmt.Errorf("Sign: %w", err)
	}

	fmt.Println("SIG:", sig)
	return nil
}
