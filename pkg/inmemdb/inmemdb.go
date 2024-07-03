package inmemdb

import (
	"errors"
	"fmt"
	"sync"
)

type operationKind = string

const (
	operationSet    operationKind = "set"
	operationDelete operationKind = "delete"
	operationGet    operationKind = "get"
)

// InMemDB is an in-memory database
type InMemDB struct {
	mutex              sync.RWMutex
	storage            map[string]string
	currentTransaction *Transaction
}

// New creates a new InMemDB
func New() *InMemDB {
	return &InMemDB{
		storage: map[string]string{},
	}
}

// CommitTransaction commits the current transaction changes
func (db *InMemDB) StartTransaction() {
	if db.currentTransaction == nil {
		db.mutex.Lock()
	}
	newTransaction := &Transaction{running: false, db: db, parent: db.currentTransaction}
	db.currentTransaction = newTransaction
}

// Set sets a key in the database
func (db *InMemDB) Set(key string, value string) (err error) {
	if key == "" {
		err = errors.New("key cannot be empty")
	}

	if db.currentTransaction != nil && !db.currentTransaction.running {
		prevValue := db.Get(key)
		newChange := change{key: key, kind: operationSet, newValue: &value, prevValue: &prevValue}
		db.currentTransaction.changes = append(db.currentTransaction.changes, newChange)
	}

	if db.currentTransaction == nil {
		db.mutex.Lock()
		defer db.mutex.Unlock()
	}

	db.unsafeSet(key, value)

	return nil
}

// unsafeSet sets a key in the database without locking
func (db *InMemDB) unsafeSet(key string, value string) {
	db.storage[key] = value
}

// Get gets a key from the database
func (db *InMemDB) Get(key string) string {
	if db.currentTransaction == nil {
		db.mutex.RLock()
		defer db.mutex.RUnlock()
	}

	// fallback and original call to the storage
	value, ok := db.storage[key]
	if !ok {
		return ""
	}

	return value
}

// Delete deletes a key from the database
func (db *InMemDB) Delete(key string) (err error) {
	value, ok := db.storage[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	if db.currentTransaction != nil && !db.currentTransaction.running {
		newChange := change{key: key, kind: operationDelete, newValue: nil, prevValue: &value}
		db.currentTransaction.changes = append(db.currentTransaction.changes, newChange)
	}

	if db.currentTransaction == nil {
		db.mutex.Lock()
		defer db.mutex.Unlock()
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
	if db.currentTransaction == nil {
		return errors.New("no transaction to commit")
	}

	if db.currentTransaction.parent == nil {
		defer db.mutex.Unlock()
	}

	return db.currentTransaction.Commit()
}

// Rollback rolls back latest transaction
func (db *InMemDB) Rollback() (err error) {
	if db.currentTransaction == nil {
		return errors.New("no transaction to rollback")
	}

	if db.currentTransaction.parent == nil {
		defer db.mutex.Unlock()
	}

	err = db.currentTransaction.Rollback()

	db.currentTransaction = db.currentTransaction.parent

	return err
}
