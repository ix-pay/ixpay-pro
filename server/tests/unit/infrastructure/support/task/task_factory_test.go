package task

import (
	"encoding/json"
	"testing"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestNewTaskFactory(t *testing.T) {
	factory := tsk.NewTaskFactory()
	assert.NotNil(t, factory)
}

func TestCreateTask_HTTP(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"url": "http://example.com", "method": "GET", "timeout": 10}`)

	createdTask, err := factory.CreateTask("http_test", tsk.TaskTypeHTTP, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "http_test", createdTask.GetName())
}

func TestCreateTask_HTTP_InvalidJSON(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`invalid`)

	createdTask, err := factory.CreateTask("http_test", tsk.TaskTypeHTTP, params)

	assert.Error(t, err)
	assert.Nil(t, createdTask)
}

func TestCreateTask_HTTP_MissingURL(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"method": "GET"}`)

	createdTask, err := factory.CreateTask("http_test", tsk.TaskTypeHTTP, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "http_test", createdTask.GetName())
}

func TestCreateTask_Database(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"query": "SELECT 1"}`)

	createdTask, err := factory.CreateTask("db_test", tsk.TaskTypeDatabase, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "db_test", createdTask.GetName())
}

func TestCreateTask_Database_MissingQuery(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{}`)

	createdTask, err := factory.CreateTask("db_test", tsk.TaskTypeDatabase, params)

	assert.Error(t, err)
	assert.Nil(t, createdTask)
}

func TestCreateTask_Cache(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"action": "delete", "cache_keys": ["key1"]}`)

	createdTask, err := factory.CreateTask("cache_test", tsk.TaskTypeCache, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "cache_test", createdTask.GetName())
}

func TestCreateTask_Script(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"command": "echo", "args": ["hello"]}`)

	createdTask, err := factory.CreateTask("script_test", tsk.TaskTypeScript, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "script_test", createdTask.GetName())
}

func TestCreateTask_Script_MissingCommand(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"args": ["hello"]}`)

	createdTask, err := factory.CreateTask("script_test", tsk.TaskTypeScript, params)

	assert.Error(t, err)
	assert.Nil(t, createdTask)
}

func TestCreateTask_UnknownType(t *testing.T) {
	factory := tsk.NewTaskFactory()

	createdTask, err := factory.CreateTask("test", "unknown_type", nil)

	assert.Error(t, err)
	assert.Nil(t, createdTask)
}

func TestCreateTask_StreamMaintenance(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"stream_key": "stream:test", "max_length": 1000}`)

	createdTask, err := factory.CreateTask("stream_test", tsk.TaskTypeStreamMaintenance, params)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "stream_test", createdTask.GetName())
}

func TestCreateTask_StreamMaintenance_MissingKey(t *testing.T) {
	factory := tsk.NewTaskFactory()

	params := json.RawMessage(`{"max_length": 1000}`)

	createdTask, err := factory.CreateTask("stream_test", tsk.TaskTypeStreamMaintenance, params)

	assert.Error(t, err)
	assert.Nil(t, createdTask)
}
