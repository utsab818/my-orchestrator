package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/utsab818/my-orchestrator/manager"
	"github.com/utsab818/my-orchestrator/task"
	"github.com/utsab818/my-orchestrator/worker"
)

func main() {
	whost := os.Getenv("WORKER_HOST")
	wport, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	mhost := os.Getenv("MANAGER_HOST")
	mport, _ := strconv.Atoi(os.Getenv("MANAGER_PORT"))

	// start api for worker
	fmt.Println("Starting my-orchestrator worker")
	w := worker.Worker{
		Queue: queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	wapi := worker.Api{
		Address: whost,
		Port:    wport,
		Worker:  &w,
	}

	go w.RunTasks()
	go w.CollectStats()
	go w.UpdateTasks()
	go wapi.Start()

	// start api for manager
	fmt.Println("Starting my-orchestrator manager")
	workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
	m := manager.New(workers)
	mapi := manager.Api{
		Address: mhost,
		Port:    mport,
		Manager: m,
	}

	go m.ProcessTasks()
	go m.UpdateTasks()
	go m.DoHealthChecks()
	mapi.Start()

}

// WORKER_HOST=localhost WORKER_PORT=5555 MANAGER_HOST=localhost MANAGER_PORT=5556 go run main.go
// curl -v -X POST localhost:5556/tasks -d @task1.json
// curl http://localhost:5556/tasks|jq
