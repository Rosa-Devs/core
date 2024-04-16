package models

type State struct {
	Db  bool
	Api bool
}

type Status struct {
	DhtNodes    int
	Connections int
	Channels    int
}

type Metrics struct {
	State  State
	Status Status
}
