package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/utsab818/my-orchestrator/task"
	"github.com/utsab818/my-orchestrator/worker"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*task.Task
	EventDb       map[uuid.UUID]*task.TaskEvent
	Workers       []string // The format could be <hostname>:<port> as we pass host and port for worker to know which worker it is.
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func New(workers []string) *Manager {
	taskDb := make(map[uuid.UUID]*task.Task)
	eventDb := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)

	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	return &Manager{
		Pending:       *queue.New(),
		Workers:       workers,
		TaskDb:        taskDb,
		EventDb:       eventDb,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
	}
}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
}

// We will first implement selection of worker for running the task
// with a naive round-robin scheduling algorithm.

// 1. Check whether we are at the end of the Workers list.
// 2. If we are not, select the next worker in the list.
// 3. Else, return to the beginning and select the first worker in the list.

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

// For each worker
// 1. Query the worker to get a list of its tasks.
// 2. For each task, update its state in the manager’s database so it
// matches the state from the worker.

func (m *Manager) UpdateTasks() {
	for _, worker := range m.Workers {
		log.Printf("Checking worker %v for task updates", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", worker, err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request: %v\n", err)
		}

		d := json.NewDecoder(resp.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Printf("Error marshalling tasks: %s\n", err.Error())
		}

		for _, t := range tasks {
			log.Printf("Attempting to update task %v\n", t.ID)
			_, ok := m.TaskDb[t.ID]
			if !ok {
				log.Printf("Task with ID %s not found\n", t.ID)
				return
			}

			if m.TaskDb[t.ID].State != t.State {
				m.TaskDb[t.ID].State = t.State
			}

			m.TaskDb[t.ID].StartTime = t.StartTime
			m.TaskDb[t.ID].FinishTime = t.FinishTime
			m.TaskDb[t.ID].ContainerId = t.ContainerId
		}
	}
}

// 1. Checks whether there are task events in the Pending queue
// 2. If there are, selects a worker to run a task
// 3. Pulls a task event off the pending queue
// 4. Sets the state of the task to Scheduled
// 5. Performs some administrative work that makes it easy for the manager to keep track of which workers tasks are running on
// 6. JSON-encodes the task event
// 7. Sends the task event to the selected worker
// 8. Checks the response from the worker

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		w := m.SelectWorker()
		e := m.Pending.Dequeue()
		te := e.(task.TaskEvent) // convert event to taskEvent type as items gets stored as the interface{} type.
		t := te.Task
		log.Printf("Pulled %v off pending queue\n", t)
		m.EventDb[te.ID] = &te
		m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], te.Task.ID)
		m.TaskWorkerMap[t.ID] = w

		t.State = task.Scheduled
		m.TaskDb[t.ID] = &t

		data, err := json.Marshal(te)
		if err != nil {
			log.Printf("Unable to marshal task object: %v.\n", t)
		}

		url := fmt.Sprintf("http://%s/tasks", w)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", w, err)
			m.Pending.Enqueue(te)
			return
		}
		d := json.NewDecoder(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			e := worker.ErrResponse{}
			err := d.Decode(&e)
			if err != nil {
				fmt.Printf("Error decoding response: %s\n", err.Error())
				return
			}
			log.Printf("Response error (%d): %s", e.HTTPStatusCode, e.Message)
			return
		}

		t = task.Task{}
		err = d.Decode(&t)
		if err != nil {
			fmt.Printf("Error decoding response: %s\n", err.Error())
			return
		}
		log.Printf("%#v\n", t)
	} else {
		log.Println("No work in the queue")
	}
}
