package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tsk "github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/stretchr/testify/assert"
)

func TestHTTPTask_Run_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ht := &tsk.HTTPTask{
		URL:            server.URL,
		Method:         "GET",
		Timeout:        5,
		ExpectedStatus: 200,
	}

	err := ht.Run(context.Background())
	assert.NoError(t, err)
}

func TestHTTPTask_Run_StatusMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ht := &tsk.HTTPTask{
		URL:            server.URL,
		Method:         "GET",
		Timeout:        5,
		ExpectedStatus: 200,
	}

	err := ht.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "状态码不匹配")
}

func TestHTTPTask_Run_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ht := &tsk.HTTPTask{
		URL:            server.URL,
		Method:         "GET",
		Timeout:        1,
		ExpectedStatus: 200,
	}

	err := ht.Run(context.Background())
	assert.Error(t, err)
}
