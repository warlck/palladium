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

	// =====================================================================
	v := struct {
		Name string
	}{
		Name: "Quincy",
	}

	data, err := stamp(v)
	if err != nil {
		return fmt.Errorf("Stamp: %w", err)
	}

	txHash := crypto.Keccak256(data)
	// =====================================================================

	// Sign the hash with the private key to produce a signature.
	sig, err := crypto.Sign(txHash, privateKey)
	if err != nil {
		return fmt.Errorf("Sign: %w", err)
	}

	fmt.Println("SIG:", sig)

	sigPublicKey, err := crypto.SigToPub(txHash, sig)
	if err != nil {
		return err
	}

	address = crypto.PubkeyToAddress(*sigPublicKey).String()
	fmt.Println("address: ", address)

	return nil
}

// stamp returns a hash of 32 bytes that represen ts this data with
// the Palladium stamp embedded into the final hash.
func stamp(value any) ([]byte, error) {

	// Marshal the data.
	v, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	// This stamp is used so signatures we produce when signing data
	// are always unique to the Palladium blockchain.
	stamp := []byte(fmt.Sprintf("\x19Palladium Signed Message:\n%d", len(v)))

	// Hash the stamp and txHash together in a final 32 byte array
	// that represents the data.
	data := crypto.Keccak256(stamp, v)

	return data, nil
}
