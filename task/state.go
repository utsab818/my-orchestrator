package task

type State int

// iota automatically increments values starting from 0.
// 	Pending = 0
// 	Scheduled = 1
// 	Running = 2
// 	Completed = 3
// 	Failed = 4

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

func (s State) String() []string {
	return []string{"Pending", "Scheduled", "Running", "Completed", "Failed"}
}

var StateTransitionMap = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func Contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func ValidStateTransition(src State, dst State) bool {
	return Contains(StateTransitionMap[src], dst)
}
