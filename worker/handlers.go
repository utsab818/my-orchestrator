package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/utsab818/my-orchestrator/task"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"http_status_code"`
	Message        string `json:"message"`
}

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Println(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Worker.AddTask(te.Task)
	log.Printf("Added task %v\n", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)
	}
	// convert uuid string to uuid.UUID type
	tID, _ := uuid.Parse(taskID)
	taskToStop, err := a.Worker.Db.Get(tID.String())
	if err != nil {
		log.Printf("No task with ID %v found", tID)
		w.WriteHeader(404)
	}
	// 	we’re using the worker’s datastore to
	// represent the current state of tasks, while we’re using the worker’s queue
	// to represent the desired state of tasks. As a result of this decision, the
	// API cannot simply retrieve the task from the worker’s datastore, set the
	// state to task.Completed, and then put the task onto the worker’s
	// queue. The reason is that the values in the datastore are pointers to
	// task.Task types. If we were to change the state on taskToStop,
	// we would be changing the state field on the task in the datastore. We
	// would then add the same task to the worker’s queue, and when it popped
	// the task off to work on it, it would complain about not being able to
	// transition a task from the state task.Completed to task.Completed.
	// Hence, we make a copy, change the state on the copy, and add it to the queue.
	taskCopy := *taskToStop.(*task.Task)
	taskCopy.State = task.Completed
	a.Worker.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskCopy.ID.String(), taskCopy.ContainerId)
	w.WriteHeader(204)
}

func (a *Api) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.Stats)
}
