package task

import (
	"context"
	"runtime"
	"testing"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestScriptTask_Run_Echo(t *testing.T) {
	var command string
	var args []string

	if runtime.GOOS == "windows" {
		command = "cmd"
		args = []string{"/C", "echo", "hello"}
	} else {
		command = "echo"
		args = []string{"hello"}
	}

	st := &tsk.ScriptTask{
		TaskID:  "script_echo",
		Command: command,
		Args:    args,
		Timeout: 10,
	}

	err := st.Run(context.Background())
	assert.NoError(t, err)
}

func TestScriptTask_Run_NonExistentCommand(t *testing.T) {
	st := &tsk.ScriptTask{
		TaskID:  "script_fail",
		Command: "non_existent_command_xyz",
		Timeout: 5,
	}

	err := st.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "脚本执行失败")
}
