package entity

import "time"

// Node 节点领域实体
type Node struct {
	NodeID        string
	Role          string
	Status        string
	IPAddress     string
	Port          int
	RunningTasks  int
	MaxConcurrent int
	LastHeartbeat time.Time
	RegisteredAt  time.Time
	StartedAt     time.Time
}
