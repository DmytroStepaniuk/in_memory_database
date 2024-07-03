package inmemdb

import (
	"errors"
	"fmt"
	"sync"

	"github.com/DmytroStepaniuk/in_memory_database/pkg/errorlist"
)

type operationKind = string

const (
	operationSet    operationKind = "set"
	operationDelete operationKind = "delete"
	operationGet    operationKind = "get"
)

// InMemDB is an in-memory database
type InMemDB struct {
	mutex       sync.RWMutex
	storage     map[string]string
	transaction *Transaction
}

type Transaction struct {
	changes []change
	running bool
}

type change struct {
	kind      operationKind
	key       string
	newValue  *string
	prevValue *string
	done      bool
}

// New creates a new InMemDB
func New() *InMemDB {
	return &InMemDB{
		storage: map[string]string{},
	}
}

// CommitTransaction commits the current transaction changes
func (db *InMemDB) StartTransaction() {
	db.transaction = &Transaction{running: false}
	db.mutex.Lock()
}

// Rollback rolls back the current transaction changes
func (db *InMemDB) Rollback() (err error) {
	defer db.mutex.Unlock()

	// clean up transaction
	// clean up transaction
	defer func() {
		db.transaction = nil
	}()

	errList := errorlist.New()

	for i := len(db.transaction.changes) - 1; i >= 0; i-- {
		change := db.transaction.changes[i]

		if !change.done {
			continue
		}

		switch change.kind {
		case operationSet:
			if change.prevValue == nil {
				err = db.Delete(change.key)
			} else {
				err = db.Set(change.key, *change.prevValue)
			}
		case operationDelete:
			err = db.Set(change.key, *change.prevValue)
		}

		if err != nil {
			errList.Add(err)
		}
	}

	return errList.Error()
}

// Set sets a key in the database
func (db *InMemDB) Set(key string, value string) (err error) {
	if key == "" {
		err = errors.New("key cannot be empty")
	}

	if db.transaction != nil && !db.transaction.running {
		newChange := change{key: key, kind: operationSet, newValue: &value, prevValue: nil, done: false}
		db.transaction.changes = append(db.transaction.changes, newChange)
		return nil
	}

	if db.transaction == nil {
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
func (db *InMemDB) Get(key string) (string, error) {
	if db.transaction == nil {
		db.mutex.RLock()
		defer db.mutex.RUnlock()
	} else {
		for i := len(db.transaction.changes) - 1; i >= 0; i-- {
			change := db.transaction.changes[i]

			if change.key == key && change.kind == operationSet {
				return *change.newValue, nil
			}

			if change.key == key && change.kind == operationDelete {
				return "", fmt.Errorf("key %s not found", key)
			}
		}
	}

	value, ok := db.storage[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}

	return value, nil
}

// Delete deletes a key from the database
func (db *InMemDB) Delete(key string) (err error) {
	value, ok := db.storage[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	if db.transaction != nil && !db.transaction.running {
		newChange := change{key: key, kind: operationDelete, newValue: nil, prevValue: &value, done: false}
		db.transaction.changes = append(db.transaction.changes, newChange)
		return nil
	}

	if db.transaction == nil {
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

// Commit commits the current transaction changes
func (db *InMemDB) Commit() (err error) {
	if db.transaction == nil {
		return errors.New("no transaction in progress")
	}

	defer db.mutex.Unlock()

	// clean up transaction
	defer func() {
		if err == nil {
			db.transaction = nil
		}
	}()

	db.transaction.running = true

	for i := 0; i < len(db.transaction.changes); i++ {
		change := db.transaction.changes[i]

		if change.done {
			continue
		}

		switch change.kind {
		case operationSet:
			err = db.Set(change.key, *change.newValue)
		case operationDelete:
			err = db.Delete(change.key)
		}

		if err == nil {
			change.done = true
		} else {
			rollbackError := db.Rollback()
			return errors.Join(rollbackError, err)
		}
	}

	return nil
}
