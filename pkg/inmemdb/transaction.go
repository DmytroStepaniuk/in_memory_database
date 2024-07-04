package inmemdb

import (
	"github.com/DmytroStepaniuk/in_memory_database/pkg/errorlist"
)

type Transaction struct {
	changes  []change
	db       *InMemDB
	finished bool
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

	for i := len(tr.changes) - 1; i >= 0; i-- {
		change := tr.changes[i]

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
	defer func() {
		tr.finished = true
	}()

	for _, change := range tr.changes {
		switch change.kind {
		case operationSet:
			err = tr.db.Set(change.key, *change.newValue)
		case operationDelete:
			err = tr.db.Delete(change.key)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *InMemDB) FindLastUnfinished() *Transaction {
	for i := len(db.transactions) - 1; i >= 0; i-- {
		if !db.transactions[i].finished {
			return db.transactions[i]
		}
	}

	return nil
}
