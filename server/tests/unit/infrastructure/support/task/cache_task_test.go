package task

import (
	"context"
	"testing"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestCacheTask_Run(t *testing.T) {
	ct := &tsk.CacheTask{
		TaskID:    "cache_run",
		Action:    "delete",
		CacheKeys: []string{"key1"},
	}

	err := ct.Run(context.Background())
	assert.NoError(t, err)
}

func TestCacheTask_Run_EmptyAction(t *testing.T) {
	ct := &tsk.CacheTask{
		TaskID: "cache_empty",
		Action: "",
	}

	err := ct.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缓存操作类型不能为空")
}
