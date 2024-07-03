package inmemdb

import (
	"github.com/DmytroStepaniuk/in_memory_database/pkg/errorlist"
)

type Transaction struct {
	changes []change
	running bool
	db      *InMemDB
	parent  *Transaction
}

type change struct {
	kind      operationKind
	key       string
	newValue  *string
	prevValue *string
}

// Rollback rolls back the current transaction changes
func (tr *Transaction) Rollback() (err error) {
	errList := errorlist.New()

	for i := len(tr.db.currentTransaction.changes) - 1; i >= 0; i-- {
		change := tr.db.currentTransaction.changes[i]

		switch change.kind {
		case operationSet:
			if change.prevValue == nil {
				err = tr.db.Delete(change.key)
			} else {
				err = tr.db.Set(change.key, *change.prevValue)
			}
		case operationDelete:
			err = tr.db.Set(change.key, *change.prevValue)
		}

		if err != nil {
			errList.Add(err)
		}
	}

	return errList.Error()
}

// Commit commits the current transaction changes
func (tr *Transaction) Commit() (err error) {
	tr.db.currentTransaction = tr.parent
	tr = nil

	return nil
}
