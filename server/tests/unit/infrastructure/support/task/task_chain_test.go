package task

import (
	"context"
	"testing"
	"time"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestTaskChainBuilder_AddTask(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("test_chain", log).
		AddTask("task1", &mockTask{name: "task1"}).
		AddTask("task2", &mockTask{name: "task2"}).
		Build()

	assert.NotNil(t, chain)
	assert.Equal(t, "test_chain", chain.GetName())
	tasks := chain.GetTasks()
	assert.Len(t, tasks, 2)
}

func TestTaskChainBuilder_AddTaskWithDeps(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("deps_chain", log).
		AddTask("task1", &mockTask{name: "task1"}).
		AddTaskWithDeps("task2", &mockTask{name: "task2"}, []string{"task1"}).
		Build()

	assert.NotNil(t, chain)
}

func TestTaskChainBuilder_EmptyChain(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("empty", log).Build()
	assert.Nil(t, chain)
}

func TestTaskChainBuilder_InvalidDependency(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("invalid_deps", log).
		AddTaskWithDeps("task1", &mockTask{name: "task1"}, []string{"non_existent"}).
		Build()

	assert.Nil(t, chain)
}

func TestTaskChainBuilder_CircularDependency(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("circular", log).
		AddTaskWithDeps("task1", &mockTask{name: "task1"}, []string{"task2"}).
		AddTaskWithDeps("task2", &mockTask{name: "task2"}, []string{"task1"}).
		Build()

	assert.Nil(t, chain)
}

func TestTaskChainBuilder_Options(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("options", log).
		OnError(tsk.ChainSkipOnError).
		Timeout(5*time.Minute).
		MaxRetries(5).
		Description("test chain").
		AddTask("task1", &mockTask{name: "task1"}).
		Build()

	assert.NotNil(t, chain)
}

func TestTaskChain_Execute_Success(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("success", log).
		AddTask("task1", &mockTask{name: "task1"}).
		AddTask("task2", &mockTask{name: "task2"}).
		Build()

	err := chain.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, tsk.ChainSuccess, chain.GetStatus())
}

func TestTaskChain_Execute_StopOnError(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("stop", log).
		OnError(tsk.ChainStopOnError).
		AddTask("task1", &mockTask{name: "task1"}).
		AddTask("task2", &mockTask{name: "failing", fail: true}).
		AddTask("task3", &mockTask{name: "task3"}).
		Build()

	err := chain.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, tsk.ChainFailed, chain.GetStatus())
}

func TestTaskChain_Execute_SkipOnError(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("skip", log).
		OnError(tsk.ChainSkipOnError).
		AddTask("task1", &mockTask{name: "task1"}).
		AddTask("task2", &mockTask{name: "failing", fail: true}).
		AddTask("task3", &mockTask{name: "task3"}).
		Build()

	err := chain.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, tsk.ChainSuccess, chain.GetStatus())
}

func TestTaskChain_GetStatus(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("status", log).
		AddTask("task1", &mockTask{name: "task1"}).
		Build()

	assert.Equal(t, tsk.ChainPending, chain.GetStatus())
}

func TestTaskChain_GetTaskStatus_NotFound(t *testing.T) {
	log := &mockLogger{}

	chain := tsk.NewTaskChain("not_found", log).
		AddTask("task1", &mockTask{name: "task1"}).
		Build()

	_, err := chain.GetTaskStatus("non_existent")
	assert.Error(t, err)
}

func TestAddTaskChainToManager(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	chain := tsk.NewTaskChain("manager_chain", log).
		AddTask("task1", &mockTask{name: "task1"}).
		Build()

	err := tm.AddTaskChain(chain)
	assert.NoError(t, err)

	tasks := tm.GetAllTasks()
	assert.Len(t, tasks, 1)
}

func TestAddTaskChainToManager_NilChain(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	err := tm.AddTaskChain(nil)
	assert.Error(t, err)
}
