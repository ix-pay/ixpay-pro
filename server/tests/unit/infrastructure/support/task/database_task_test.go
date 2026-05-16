package task

import (
	"context"
	"testing"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseTask_Run(t *testing.T) {
	dt := &tsk.DatabaseTask{
		TaskID: "db_run",
		Query:  "SELECT 1",
	}

	err := dt.Run(context.Background())
	assert.NoError(t, err)
}

func TestDatabaseTask_Run_EmptyQuery(t *testing.T) {
	dt := &tsk.DatabaseTask{
		TaskID: "db_empty",
		Query:  "",
	}

	err := dt.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SQL 查询为空")
}
