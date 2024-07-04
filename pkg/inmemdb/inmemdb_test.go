package inmemdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DmytroStepaniuk/in_memory_database/pkg/inmemdb"
)

func TestInMemDB(t *testing.T) {
	var err error

	// # Example 1 for commit a transaction
	// db = InMemoryDatabase()
	// db.set("key1", "value1")
	// db.start_transaction()
	// db.set("key1", "value2")
	// db.commit()
	// db.get(“key1”)    -> Expect to get “value2”
	t.Run("Example 1", func(t *testing.T) {
		inMemDb := inmemdb.New()

		err = inMemDb.Set("key1", "value1")
		assert.Nil(t, err)

		inMemDb.StartTransaction()
		err = inMemDb.Set("key1", "value2")
		assert.Nil(t, err)

		err = inMemDb.Commit()
		assert.Nil(t, err)

		value := inMemDb.Get("key1")
		assert.Equal(t, "value2", value)
	})

	// # Example 2 for roll_back().
	// db = InMemoryDatabase()
	// db.set("key1", "value1")
	// db.start_transaction()
	// db.get("key1")    -> Expect to get “value1”
	// db.set("key1", "value2")
	// db.get("key1")    -> Expect to get ”value2”
	// db.roll_back()
	// db.get(“key1”)    -> Expect to get “value1”
	t.Run("Example 2", func(t *testing.T) {
		inMemDb := inmemdb.New()

		err = inMemDb.Set("key1", "value1")
		assert.Nil(t, err)

		inMemDb.StartTransaction()

		value := inMemDb.Get("key1")
		assert.Equal(t, "value1", value)

		err = inMemDb.Set("key1", "value2")
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "value2", value)

		err = inMemDb.Rollback()
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "value1", value)
	})

	// # Example 3 for nested transactions
	// db = InMemoryDatabase()
	// db.set("key1", "value1")
	// db.start_transaction()
	// db.set("key1", "value2")
	// db.get("key1")    -> Expect to get ”value2”
	// db.start_transaction()
	// db.get("key1")    -> Expect to get ”value2”
	// db.delete(“key1“)
	// db.commit()
	// db.get(“key1”)    -> Expect to get None
	// db.commit()
	// db.get(“key1”)     -> Expect to get None
	t.Run("Example 3", func(t *testing.T) {
		inMemDb := inmemdb.New()

		err = inMemDb.Set("key1", "value1")
		assert.Nil(t, err)

		inMemDb.StartTransaction()
		err = inMemDb.Set("key1", "value2")
		assert.Nil(t, err)

		value := inMemDb.Get("key1")
		assert.Equal(t, "value2", value)

		inMemDb.StartTransaction()

		value = inMemDb.Get("key1")
		assert.Equal(t, "value2", value)

		err = inMemDb.Delete("key1")
		assert.Nil(t, err)

		err = inMemDb.Commit()
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "", value)

		err = inMemDb.Commit()
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "", value)
	})

	// # Example 4 for nested transactions with roll_back()
	// db = InMemoryDatabase()
	// db.set("key1", "value1")
	// db.start_transaction()
	// db.set("key1", "value2")
	// db.get("key1")    -> Expect to get ”value2”
	// db.start_transaction()
	// db.get("key1")    -> Expect to get ”value2”
	// db.delete(“key1“)
	// db.roll_back()
	// db.get(“key1”)    -> Expect to get “value2”
	// db.commit()
	// db.get(“key1”)     -> Expect to get “value2”
	t.Run("Example 4", func(t *testing.T) {
		inMemDb := inmemdb.New()

		err = inMemDb.Set("key1", "value1")
		assert.Nil(t, err)

		inMemDb.StartTransaction()

		err = inMemDb.Set("key1", "value2")
		assert.Nil(t, err)

		value := inMemDb.Get("key1")
		assert.Equal(t, "value2", value)

		inMemDb.StartTransaction()

		value = inMemDb.Get("key1")
		assert.Equal(t, "value2", value)

		err = inMemDb.Delete("key1")
		assert.Nil(t, err)

		err = inMemDb.Rollback()
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "value2", value)

		err = inMemDb.Commit()
		assert.Nil(t, err)

		value = inMemDb.Get("key1")
		assert.Nil(t, err)
		assert.Equal(t, "value2", value)
	})
}
