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
	// privateKey, err := crypto.GenerateKey()
	// if err != nil {
	// 	return err
	// }

	// crypto.SaveECDSA("zblock/accounts/quincy.ecdsa", privateKey)

	// Need to load the private key file for the configured beneficiary so the
	// account can get credited with fees and tips.
	path := fmt.Sprintf("%s%s.ecdsa", "zblock/accounts/", "quincy")
	privateKey, err := crypto.LoadECDSA(path)
	if err != nil {
		return fmt.Errorf("unable to load private key for node: %w", err)
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
