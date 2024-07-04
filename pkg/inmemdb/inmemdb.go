package inmemdb

import (
	"errors"
	"fmt"
)

type operationKind = string

const (
	operationSet    operationKind = "set"
	operationDelete operationKind = "delete"
	operationGet    operationKind = "get"
)

// InMemDB is an in-memory database
type InMemDB struct {
	storage      map[string]string
	transactions []*Transaction
}

// New creates a new InMemDB
func New() *InMemDB {
	return &InMemDB{
		storage: map[string]string{},
	}
}

// CommitTransaction commits the current transaction changes
func (db *InMemDB) StartTransaction() {
	newTransaction := &Transaction{db: db}
	db.transactions = append(db.transactions, newTransaction)
}

func (db *InMemDB) AnyTransactions() bool {
	return len(db.transactions) > 0
}

// FindLastUnfinished returns the last transaction
func (db *InMemDB) lastTransaction() *Transaction {
	if len(db.transactions) == 0 {
		return nil
	}

	return db.transactions[len(db.transactions)-1]
}

// Set sets a key in the database
func (db *InMemDB) Set(key string, value string) (err error) {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	if db.AnyTransactions() {
		prevValue := db.Get(key)
		newChange := change{key: key, kind: operationSet, newValue: &value, prevValue: prevValue}
		db.FindLastUnfinished().changes = append(db.FindLastUnfinished().changes, newChange)
		return nil
	}

	db.unsafeSet(key, value)

	return nil
}

// unsafeSet sets a key in the database without locking
func (db *InMemDB) unsafeSet(key string, value string) {
	db.storage[key] = value
}

// Get gets a key from the database
func (db *InMemDB) Get(key string) *string {
	if db.AnyTransactions() {
		for i := len(db.transactions) - 1; i >= 0; i-- {
			current := db.transactions[i]

			for _, change := range current.changes {
				if change.key == key && change.kind == operationSet {
					return change.newValue
				}

				if change.key == key && change.kind == operationDelete {
					return nil
				}
			}
		}
	}

	// fallback and original call to the storage
	value, ok := db.storage[key]
	if !ok {
		return nil
	}

	return &value
}

// Delete deletes a key from the database
func (db *InMemDB) Delete(key string) (err error) {
	value, ok := db.storage[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	if db.AnyTransactions() {
		newChange := change{key: key, kind: operationDelete, newValue: nil, prevValue: &value}
		db.lastTransaction().changes = append(db.lastTransaction().changes, newChange)
		return nil
	}

	db.unsafeDelete(key)

	return nil
}

// unsafeDelete deletes a key from the database without locking
func (db *InMemDB) unsafeDelete(key string) {
	delete(db.storage, key)
}

// Commit commits all transactions
func (db *InMemDB) Commit() (err error) {
	if !db.AnyTransactions() {
		return errors.New("no transaction to commit")
	}
	err = db.FindLastUnfinished().Commit()
	return err
}

// Rollback rolls back latest transaction
func (db *InMemDB) Rollback() error {
	if !db.AnyTransactions() {
		return errors.New("no transaction to rollback")
	}

	err := db.FindLastUnfinished().Rollback()

	db.FindLastUnfinished().changes = []change{}

	return err
}
