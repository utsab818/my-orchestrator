package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/utsab818/my-orchestrator/manager"
	"github.com/utsab818/my-orchestrator/worker"
)

func main() {
	whost := os.Getenv("WORKER_HOST")
	wport, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	mhost := os.Getenv("MANAGER_HOST")
	mport, _ := strconv.Atoi(os.Getenv("MANAGER_PORT"))

	// start api for worker
	fmt.Println("Starting my-orchestrator worker")
	// w1 := worker.New("worker-1", "memory")
	// w2 := worker.New("worker-2", "memory")
	// w3 := worker.New("worker-3", "memory")

	w1 := worker.New("worker-1", "persistent")
	w2 := worker.New("worker-2", "persistent")
	w3 := worker.New("worker-3", "persistent")

	wapi1 := worker.Api{
		Address: whost,
		Port:    wport,
		Worker:  w1,
	}

	wapi2 := worker.Api{
		Address: whost,
		Port:    wport + 1,
		Worker:  w2,
	}

	wapi3 := worker.Api{
		Address: whost,
		Port:    wport + 2,
		Worker:  w3,
	}

	go w1.RunTasks()
	go w1.UpdateTasks()
	go wapi1.Start()

	go w2.RunTasks()
	go w2.UpdateTasks()
	go wapi2.Start()

	go w3.RunTasks()
	go w3.UpdateTasks()
	go wapi3.Start()

	// start api for manager
	fmt.Println("Starting my-orchestrator manager")
	workers := []string{
		fmt.Sprintf("%s:%d", whost, wport),
		fmt.Sprintf("%s:%d", whost, wport+1),
		fmt.Sprintf("%s:%d", whost, wport+2),
	}

	// m := manager.New(workers, "roundrobin")
	// m := manager.New(workers, "epvm", "memory")
	m := manager.New(workers, "epvm", "persistent")

	mapi := manager.Api{
		Address: mhost,
		Port:    mport,
		Manager: m,
	}

	go m.ProcessTasks()
	go m.UpdateTasks()
	mapi.Start()

}

// WORKER_HOST=localhost WORKER_PORT=5556 MANAGER_HOST=localhost MANAGER_PORT=5555 go run main.go
// curl -v -X POST localhost:5555/tasks -d @task1.json
// curl -v --request DELETE 'localhost:5555/tasks/bb1d59ef-9fc1-4e4b-a44d-db571eeed203'
