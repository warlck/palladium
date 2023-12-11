package state

import "github.com/warlck/palladium/foundation/blockchain/database"

// QueryAccount returns a copy of the account from the database.
func (s *State) QueryAccount(account database.AccountID) (database.Account, error) {
	return s.db.Query(account)
}
