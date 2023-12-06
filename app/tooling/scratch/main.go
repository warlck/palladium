package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/warlck/palladium/foundation/blockchain/database"
	"github.com/warlck/palladium/foundation/blockchain/merkle"
	"github.com/warlck/palladium/foundation/blockchain/signature"
	"github.com/warlck/palladium/foundation/blockchain/storage/disk"
)

func main() {
	// if err := readBlock(); err != nil {
	// 	log.Fatalln(err)
	// }
	// if err := writeScratchBlock(); err != nil {
	// 	log.Fatalln(err)
	// }

	run()
}

func run() error {
	// privateKey, err := crypto.GenerateKey()
	// if err != nil {
	// 	return err
	// }

	// crypto.SaveECDSA("zblock/accounts/quincy.ecdsa", privateKey)

	// Need to load the private key file for the configured beneficiary so the
	// account can get credited with fees and tips.
	path := fmt.Sprintf("%s%s.ecdsa", "zblock/accounts/", "adam")
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

func readBlock() error {
	d, err := disk.New("zblock/testminer")
	if err != nil {
		return err
	}

	blockData, err := d.GetBlock(1)
	if err != nil {
		return err
	}
	fmt.Println(blockData)

	block, err := database.ToBlock(blockData)
	if err != nil {
		return err
	}

	if blockData.Header.TrxRoot != block.MerkleTree.RootHex() {
		return errors.New("merkle tree wrong")
	}

	fmt.Println("Merkle tree matches")

	return nil
}

func writeScratchBlock() error {
	txs := []database.Tx{
		{
			ChainID: 1,
			Nonce:   1,
			ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
			Value:   100,
			Tip:     50,
		},
		{
			ChainID: 1,
			Nonce:   2,
			ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
			Value:   100,
			Tip:     50,
		},
	}

	blockTxns := make([]database.BlockTx, len(txs))
	for i, txn := range txs {
		blockTx, err := signToBlockTxn(txn, 15)
		if err != nil {
			return err
		}
		blockTxns[i] = blockTx
	}

	tree, err := merkle.NewTree(blockTxns)
	if err != nil {
		return err
	}

	beneficiaryID, err := database.ToAccountID("0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA")
	if err != nil {
		return err
	}
	// Construct the block to be written on the disk.
	block := database.Block{
		Header: database.BlockHeader{
			Number:        1,
			PrevBlockHash: signature.ZeroHash,
			TimeStamp:     uint64(time.Now().UTC().UnixMilli()),
			BeneficiaryID: beneficiaryID,
			Difficulty:    6,
			MiningReward:  700,
			StateRoot:     "nil",
			TrxRoot:       tree.RootHex(), //
			Nonce:         0,
		},
		MerkleTree: tree,
	}

	blockData := database.NewBlockData(block)
	d, err := disk.New("zblock/testminer")
	if err != nil {
		return err
	}

	if err := d.Write(blockData); err != nil {
		return err
	}

	return nil
}

func signToBlockTxn(tx database.Tx, gas uint64) (database.BlockTx, error) {
	pk, err := crypto.HexToECDSA("fb07a21977afeb28a23ab6734ee83208af580c6ac6e0a7c1b31e28054a56a270")
	if err != nil {
		return database.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return database.BlockTx{}, err
	}

	return database.NewBlockTx(signedTx, gas, 1), nil
}
