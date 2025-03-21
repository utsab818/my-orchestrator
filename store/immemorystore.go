package store

import (
	"fmt"

	"github.com/utsab818/my-orchestrator/task"
)

// **************************************************

type InMemoryTaskStore struct {
	Db map[string]*task.Task
}

type InMemoryTaskEventStore struct {
	Db map[string]*task.TaskEvent
}

func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		Db: make(map[string]*task.Task),
	}
}

func NewInMemoryTaskEventStore() *InMemoryTaskEventStore {
	return &InMemoryTaskEventStore{
		Db: map[string]*task.TaskEvent{},
	}
}

// **************************************************

// for InMemoryTaskStore

func (i *InMemoryTaskStore) Put(key string, value any) error {
	t, ok := value.(*task.Task)
	if !ok {
		return fmt.Errorf("value %v is not a task.Task type", value)
	}
	i.Db[key] = t
	return nil
}

func (i *InMemoryTaskStore) Get(key string) (any, error) {
	t, ok := i.Db[key]
	if !ok {
		return nil, fmt.Errorf("task with key %s does not exist", key)
	}
	return t, nil
}

func (i *InMemoryTaskStore) List() (any, error) {
	var tasks []*task.Task
	for _, t := range i.Db {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (i *InMemoryTaskStore) Count() (int, error) {
	return len(i.Db), nil
}

// for InMemoryTaskEventStore

func (i *InMemoryTaskEventStore) Put(key string, value any) error {
	e, ok := value.(*task.TaskEvent)
	if !ok {
		return fmt.Errorf("value %v is not a task.TaskEvent type", value)
	}

	i.Db[key] = e
	return nil
}

func (i *InMemoryTaskEventStore) Get(key string) (any, error) {
	e, ok := i.Db[key]
	if !ok {
		return nil, fmt.Errorf("task event with key %s does not exist", key)
	}
	return e, nil
}

func (i *InMemoryTaskEventStore) List() (any, error) {
	var events []*task.TaskEvent
	for _, e := range i.Db {
		events = append(events, e)
	}
	return events, nil
}

func (i *InMemoryTaskEventStore) Count() (int, error) {
	return len(i.Db), nil
}
