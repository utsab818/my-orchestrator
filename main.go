package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/utsab818/my-orchestrator/task"
	"github.com/utsab818/my-orchestrator/worker"
)

func main() {
	host := os.Getenv("WORKER_HOST")
	port, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	fmt.Println("Starting my-orchestrator worker")
	w := worker.Worker{
		Queue: queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	api := worker.Api{
		Address: host,
		Port:    port,
		Worker:  &w,
	}

	go runTasks(&w)
	api.Start()
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Printf("No tasks to process currently.\n")
		}
		log.Println("Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}
}

// WORKER_HOST=localhost WORKER_PORT=5555 go run main.go
// curl -v localhost:5555/tasks (in next tab) --> for now provides empty list

// make post request
// curl -v --request POST \
// --header "Content-Type: application/json" \
// --data '{
//     "ID": "266592cd-960d-4091-981c-8c25c44b1018",
//     "State": 2,
//     "Task": {
//         "State": 1,
//         "ID": "266592cd-960d-4091-981c-8c25c44b1018",
//         "Name": "test",
//         "Image": "strm/helloworld-http"
//     }
// }' http://localhost:5555/tasks

// delete
// curl -v --request DELELTE "localhost:5555/tasks/266592cd-960d-4091-981c-8c25c44b1018"
