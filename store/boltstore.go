package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/utsab818/my-orchestrator/task"
)

// **************************************************

// a pure go based key-value datastore
// boltdb uses a file on disk to persist data
type TaskStore struct {
	Db       *bolt.DB
	DbFile   string      // file name to persist data on
	FileMode os.FileMode // necessary permissions for the file
	Bucket   string      // key-value pairs are store in collections called buckets
}

type EventStore struct {
	Db       *bolt.DB
	DbFile   string
	FileMode os.FileMode
	Bucket   string
}

func NewTaskStore(file string, mode os.FileMode, bucket string) (*TaskStore, error) {
	db, err := bolt.Open(file, mode, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to open %v", file)
	}

	t := TaskStore{
		DbFile:   file,
		FileMode: mode,
		Db:       db,
		Bucket:   bucket,
	}

	err = t.CreateBucket()
	if err != nil {
		log.Printf("bucket already exists, will use it instead of creating new one")
	}
	return &t, nil
}

func NewEventStore(file string, mode os.FileMode, bucket string) (*EventStore, error) {
	db, err := bolt.Open(file, mode, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to open %v", file)
	}

	e := EventStore{
		DbFile:   file,
		FileMode: mode,
		Db:       db,
		Bucket:   bucket,
	}

	err = e.CreateBucket()
	if err != nil {
		log.Printf("bucket already exists, will use it instead of creating new one")
	}

	return &e, nil
}

// **************************************************

// **************************************************

func (t *TaskStore) CreateBucket() error {
	return t.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(t.Bucket))
		if err != nil {
			return fmt.Errorf("create bucket %s: %s", t.Bucket, err)
		}
		return nil
	})

}

func (e *EventStore) CreateBucket() error {
	return e.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(e.Bucket))
		if err != nil {
			return fmt.Errorf("create bucket %s: %s", e.Bucket, err)
		}
		return nil
	})

}

func (t *TaskStore) Close() {
	t.Db.Close()
}

func (e *EventStore) Close() {
	e.Db.Close()
}

// **************************************************

// bolt supports three types of transactions.
// 1. Read-Write
// 2. Read-Only
// 3. Batch Read-Write

// For TaskStore

// For count we will be using Read-Only
func (t *TaskStore) Count() (int, error) {
	taskCount := 0
	err := t.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		b.ForEach(func(k, v []byte) error {
			taskCount++
			return nil
		})
		return nil
	})

	if err != nil {
		return -1, err
	}

	return taskCount, nil
}

// For put method we will be using Read-Write transaction
func (t *TaskStore) Put(key string, value any) error {
	return t.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		buf, err := json.Marshal(value.(*task.Task))
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), buf)
		if err != nil {
			return nil
		}

		return nil
	})
}

// For get method we will use Read-Only transaction
func (t *TaskStore) Get(key string) (any, error) {
	var task task.Task
	err := t.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		t := b.Get([]byte(key))
		if t == nil {
			return fmt.Errorf("task %v not found", key)
		}
		// decode slice of bytes to task.Task type.
		err := json.Unmarshal(t, &task)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// For list method we will use Read-Only transaction
func (t *TaskStore) List() (any, error) {
	var tasks []*task.Task
	err := t.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		b.ForEach(func(k, v []byte) error {
			var task task.Task
			err := json.Unmarshal(v, &task)
			if err != nil {
				return nil
			}
			tasks = append(tasks, &task)
			return nil
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, err
}

// For EventStore

func (e *EventStore) Count() (int, error) {
	eventCount := 0
	err := e.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(e.Bucket))
		b.ForEach(func(k, v []byte) error {
			eventCount++
			return nil
		})
		return nil
	})
	if err != nil {
		return -1, err
	}
	return eventCount, nil
}

func (e *EventStore) Put(key string, value interface{}) error {
	return e.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(e.Bucket))
		buf, err := json.Marshal(value.(*task.TaskEvent))
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), buf)
		if err != nil {
			log.Printf("unable to save item %s", key)
			return err
		}
		return nil
	})
}

func (e *EventStore) Get(key string) (interface{}, error) {
	var event task.TaskEvent
	err := e.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(e.Bucket))
		t := b.Get([]byte(key))
		if t == nil {
			return fmt.Errorf("event %v not found", key)
		}
		err := json.Unmarshal(t, &event)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (e *EventStore) List() (interface{}, error) {
	var events []*task.TaskEvent
	err := e.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(e.Bucket))
		b.ForEach(func(k, v []byte) error {
			var event task.TaskEvent
			err := json.Unmarshal(v, &event)
			if err != nil {
				return err
			}
			events = append(events, &event)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}
