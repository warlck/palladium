// Package disk implements the ability to read and write blocks to disk
// writing each block to a separate block numbered file.
package disk

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/warlck/palladium/foundation/blockchain/database"
)

// Disk represents the serialization implementation for reading and storing
// blocks in their own separate files on disk. This implements the database.Storage
// interface.
type Disk struct {
	dbPath string
}

// New constructs an Disk value for use.
func New(dbPath string) (*Disk, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	return &Disk{dbPath: dbPath}, nil
}

// Close in this implementation has nothing to do since a new file is
// written to disk for each new block and then immediately closed.
func (d *Disk) Close() error {
	return nil
}

// Write takes the specified database blocks and stores it on disk in a
// file labeled with the block number.
func (d *Disk) Write(blockData database.BlockData) error {

	// Marshal the block for writing to disk in a more human readable format.
	data, err := json.MarshalIndent(blockData, "", "  ")
	if err != nil {
		return err
	}

	// Create a new file for this block and name it based on the block number.
	f, err := os.OpenFile(d.getPath(blockData.Header.Number), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the new block to disk.
	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// GetBlock searches the blockchain on disk to locate and return the
// contents of the specified block by number.
func (d *Disk) GetBlock(num uint64) (database.BlockData, error) {

	// Open the block file for the specified number.
	f, err := os.OpenFile(d.getPath(num), os.O_RDONLY, 0600)
	if err != nil {
		return database.BlockData{}, err
	}
	defer f.Close()

	// Decode the contents of the block.
	var blockData database.BlockData
	if err := json.NewDecoder(f).Decode(&blockData); err != nil {
		return database.BlockData{}, err
	}

	// Return the block as a database block.
	return blockData, nil
}

// getPath forms the path to the specified block.
func (d *Disk) getPath(blockNum uint64) string {
	name := strconv.FormatUint(blockNum, 10)
	return path.Join(d.dbPath, fmt.Sprintf("%s.json", name))
}
