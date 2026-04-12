package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	registerURL     = "http://localhost:8081/api//register"     // 服务注册接口
	testURL         = "http://localhost:8081/test-service/test" // 正确的服务请求格式: /service-name/path
	concurrency     = 50                                        // 并发数
	totalRequests   = 10000                                     // 总请求数
	registerAuthKey = "ihvke@2025"                              // 注册认证密钥
)

// ServiceRegisterRequest 服务注册请求结构
type ServiceRegisterRequest struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
}

// 启动一个简单的测试服务
func startTestService() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"test response"}`))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	go func() {
		fmt.Println("Starting test service on :8000")
		if err := http.ListenAndServe(":8000", nil); err != nil {
			fmt.Printf("Test service failed: %v\n", err)
		}
	}()

	// 等待服务启动
	time.Sleep(500 * time.Millisecond)
}

// 注册服务
func registerService() error {
	reqBody := ServiceRegisterRequest{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8000,
		Metadata: map[string]string{
			"version":     "1.0.0",
			"health_path": "/health",
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", registerURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+registerAuthKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service: %d", resp.StatusCode)
	}

	fmt.Println("Test service registered successfully")
	return nil
}

func main() {
	fmt.Printf("Starting performance test for %s\n", testURL)
	fmt.Printf("Concurrency: %d, Total Requests: %d\n\n", concurrency, totalRequests)

	// 1. 启动测试服务
	startTestService()

	// 2. 注册测试服务
	if err := registerService(); err != nil {
		fmt.Printf("Failed to register test service: %v\n", err)
		return
	}

	// 3. 等待服务注册完成
	time.Sleep(1 * time.Second)

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 统计变量
	var totalLatency time.Duration
	var successCount int
	var failureCount int
	var mu sync.Mutex

	// 计时开始
	startTime := time.Now()

	// 启动并发请求
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			// 每个goroutine处理的请求数
			reqsPerGoroutine := totalRequests / concurrency
			if i < totalRequests%concurrency {
				reqsPerGoroutine++
			}

			for j := 0; j < reqsPerGoroutine; j++ {
				// 发送请求并计时
				reqStartTime := time.Now()
				resp, err := http.Get(testURL)
				reqLatency := time.Since(reqStartTime)

				mu.Lock()
				totalLatency += reqLatency
				mu.Unlock()

				if err != nil {
					mu.Lock()
					failureCount++
					mu.Unlock()
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					mu.Lock()
					successCount++
					mu.Unlock()
				} else {
					mu.Lock()
					failureCount++
					mu.Unlock()
				}
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()

	// 计算总耗时
	totalTime := time.Since(startTime)

	// 计算性能指标
	throughput := float64(totalRequests) / totalTime.Seconds()
	averageLatency := totalLatency / time.Duration(totalRequests)
	successRate := float64(successCount) / float64(totalRequests) * 100

	// 输出测试结果
	fmt.Printf("Test Results:\n")
	fmt.Printf("- Total Time: %.2fs\n", totalTime.Seconds())
	fmt.Printf("- Throughput: %.2f requests/second\n", throughput)
	fmt.Printf("- Average Latency: %v\n", averageLatency)
	fmt.Printf("- Success Rate: %.2f%% (%d/%d)\n", successRate, successCount, totalRequests)
	fmt.Printf("- Failure Count: %d\n", failureCount)
}
