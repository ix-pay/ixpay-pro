package task

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

type mockTask struct {
	name     string
	delay    time.Duration
	fail     bool
	runCount int
	mux      sync.Mutex
}

func (m *mockTask) Run(ctx context.Context) error {
	m.mux.Lock()
	m.runCount++
	m.mux.Unlock()
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if m.fail {
		return assert.AnError
	}
	return nil
}

func (m *mockTask) GetName() string {
	return m.name
}

func (m *mockTask) GetRunCount() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.runCount
}

type mockLogger struct {
	InfoLogs  []string
	ErrorLogs []string
	mux       sync.Mutex
}

func (m *mockLogger) Debug(msg string, fields ...interface{})               {}
func (m *mockLogger) Info(msg string, fields ...interface{}) {
	m.mux.Lock(); defer m.mux.Unlock(); m.InfoLogs = append(m.InfoLogs, msg)
}
func (m *mockLogger) Warn(msg string, fields ...interface{})               {}
func (m *mockLogger) Error(msg string, fields ...interface{}) {
	m.mux.Lock(); defer m.mux.Unlock(); m.ErrorLogs = append(m.ErrorLogs, msg)
}
func (m *mockLogger) Fatal(msg string, fields ...interface{})              {}
func (m *mockLogger) With(fields ...interface{}) logger.Logger             { return m }
func (m *mockLogger) WithContext(ctx context.Context) logger.Logger        { return m }
func (m *mockLogger) Sync() error                                          { return nil }

func TestNewTaskManager(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	assert.NotNil(t, tm)
}

func TestAddScheduledTask(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	err := tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "test_task"},
		CronExpr: "0 */1 * * * *",
		Group:    "test",
	})

	assert.NoError(t, err)
	tasks := tm.GetAllTasks()
	assert.Len(t, tasks, 1)
}

func TestAddScheduledTask_InvalidCron(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	err := tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "test_task"},
		CronExpr: "invalid",
	})

	assert.Error(t, err)
}

func TestAddScheduledTask_NilTask(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	tm.AddScheduledTask(nil)
	t.Error("Expected panic when passing nil task")
}

func TestRemoveScheduledTask(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "removable"},
		CronExpr: "0 */1 * * * *",
	})

	err := tm.RemoveScheduledTask("removable")
	assert.NoError(t, err)

	_, exists := tm.GetTask("removable")
	assert.False(t, exists)
}

func TestGetTask(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:        &mockTask{name: "info"},
		CronExpr:    "0 */1 * * * *",
		Group:       "group1",
		Concurrency: tsk.ConcurrencySkip,
	})

	info, exists := tm.GetTask("info")
	assert.True(t, exists)
	assert.Equal(t, "0 */1 * * * *", info.CronExpr)
	assert.Equal(t, "group1", info.Group)
}

func TestGetTask_NotFound(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	_, exists := tm.GetTask("non_existent")
	assert.False(t, exists)
}

func TestGetAllTasks(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "task1"},
		CronExpr: "0 */1 * * * *",
	})
	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "task2"},
		CronExpr: "0 */2 * * * *",
	})

	tasks := tm.GetAllTasks()
	assert.Len(t, tasks, 2)
}

func TestGetTasksByGroup(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "t1"},
		CronExpr: "0 */1 * * * *",
		Group:    "g1",
	})
	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "t2"},
		CronExpr: "0 */2 * * * *",
		Group:    "g1",
	})
	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "t3"},
		CronExpr: "0 */3 * * * *",
		Group:    "g2",
	})

	tasks := tm.GetTasksByGroup("g1")
	assert.Len(t, tasks, 2)
}

func TestSetTaskGroup(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "regroup"},
		CronExpr: "0 */1 * * * *",
		Group:    "old_group",
	})

	err := tm.SetTaskGroup("regroup", "new_group")
	assert.NoError(t, err)

	group, _ := tm.GetTaskGroup("regroup")
	assert.Equal(t, "new_group", group)
}

func TestStop(t *testing.T) {
	log := &mockLogger{}
	tm := tsk.SetupTaskManager(log)

	tm.AddScheduledTask(&tsk.ScheduledTask{
		Task:     &mockTask{name: "stop"},
		CronExpr: "0 */1 * * * *",
	})

	tm.Stop()
}
