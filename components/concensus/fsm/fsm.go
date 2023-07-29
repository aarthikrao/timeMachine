package fsm

type NodeConfig interface {
	// Returns the last updated time
	GetLastUpdatedTime() int

	SetChangeHandler(func() error)
}
