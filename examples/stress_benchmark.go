// Package examples
// 排行榜系统压力测试
// Author: HHaou
// Created: 2024-01-20

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// TestConfig 测试配置
type TestConfig struct {
	BaseURL     string
	Concurrency int
	Duration    time.Duration
	Timeout     time.Duration
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalTime       time.Duration
	MinTime         time.Duration
	MaxTime         time.Duration
	AvgTime         time.Duration
	QPS             float64
	SuccessRate     float64
	Errors          map[string]int64
}

// StressTest 压力测试器
type StressTest struct {
	config *TestConfig
	client *http.Client
	result *TestResult
	mutex  sync.RWMutex
}

// NewStressTest 创建压力测试器
func NewStressTest(config *TestConfig) *StressTest {
	return &StressTest{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		result: &TestResult{
			Errors:  make(map[string]int64),
			MinTime: time.Hour, // 初始化为一个大值
		},
	}
}

// makeRequest 发送HTTP请求
func (st *StressTest) makeRequest(method, path string, body interface{}) (*http.Response, time.Duration, error) {
	start := time.Now()
	
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, st.config.BaseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := st.client.Do(req)
	duration := time.Since(start)
	
	return resp, duration, err
}

// recordResult 记录测试结果
func (st *StressTest) recordResult(duration time.Duration, success bool, errorMsg string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	atomic.AddInt64(&st.result.TotalRequests, 1)
	
	if success {
		atomic.AddInt64(&st.result.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&st.result.FailedRequests, 1)
		if errorMsg != "" {
			st.result.Errors[errorMsg]++
		}
	}

	// 更新时间统计
	if duration < st.result.MinTime {
		st.result.MinTime = duration
	}
	if duration > st.result.MaxTime {
		st.result.MaxTime = duration
	}
}

// testHealthCheck 健康检查压力测试
func (st *StressTest) testHealthCheck(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	for {
		select {
		case <-stopCh:
			return
		default:
			resp, duration, err := st.makeRequest("GET", "/health", nil)
			if err != nil {
				st.recordResult(duration, false, err.Error())
				continue
			}
			
			success := resp.StatusCode == 200
			resp.Body.Close()
			
			var errorMsg string
			if !success {
				errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			
			st.recordResult(duration, success, errorMsg)
			time.Sleep(10 * time.Millisecond) // 避免过于频繁的请求
		}
	}
}

// testCreateLeaderboard 创建排行榜压力测试
func (st *StressTest) testCreateLeaderboard(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	counter := 0
	for {
		select {
		case <-stopCh:
			return
		default:
			counter++
			leaderboard := map[string]interface{}{
				"name":        fmt.Sprintf("压力测试排行榜_%d_%d", time.Now().Unix(), counter),
				"description": "压力测试创建的排行榜",
				"game_id":     fmt.Sprintf("stress_test_game_%d", rand.Intn(100)),
				"type":        "score",
				"order":       "desc",
				"max_entries": 1000,
			}
			
			resp, duration, err := st.makeRequest("POST", "/api/v1/leaderboard/create", leaderboard)
			if err != nil {
				st.recordResult(duration, false, err.Error())
				continue
			}
			
			success := resp.StatusCode == 200 || resp.StatusCode == 201
			resp.Body.Close()
			
			var errorMsg string
			if !success {
				errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			
			st.recordResult(duration, success, errorMsg)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// testSubmitScore 提交分数压力测试
func (st *StressTest) testSubmitScore(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	counter := 0
	for {
		select {
		case <-stopCh:
			return
		default:
			counter++
			score := map[string]interface{}{
				"user_id":        fmt.Sprintf("stress_user_%d", rand.Intn(1000)),
				"username":       fmt.Sprintf("压力测试用户_%d", counter),
				"score":          rand.Intn(10000),
				"leaderboard_id": "stress_test_leaderboard",
			}
			
			resp, duration, err := st.makeRequest("POST", "/api/v1/score/submit", score)
			if err != nil {
				st.recordResult(duration, false, err.Error())
				continue
			}
			
			success := resp.StatusCode == 200 || resp.StatusCode == 201
			resp.Body.Close()
			
			var errorMsg string
			if !success {
				errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			
			st.recordResult(duration, success, errorMsg)
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// testGetLeaderboard 查询排行榜压力测试
func (st *StressTest) testGetLeaderboard(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	for {
		select {
		case <-stopCh:
			return
		default:
			resp, duration, err := st.makeRequest("GET", "/api/v1/leaderboard/list", nil)
			if err != nil {
				st.recordResult(duration, false, err.Error())
				continue
			}
			
			success := resp.StatusCode == 200
			resp.Body.Close()
			
			var errorMsg string
			if !success {
				errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			
			st.recordResult(duration, success, errorMsg)
			time.Sleep(30 * time.Millisecond)
		}
	}
}

// runTest 运行压力测试
func (st *StressTest) runTest(testName string, testFunc func(*sync.WaitGroup, <-chan struct{})) {
	fmt.Printf("\n=== %s 压力测试 ===\n", testName)
	fmt.Printf("并发数: %d, 持续时间: %v\n", st.config.Concurrency, st.config.Duration)
	
	// 重置结果
	st.result = &TestResult{
		Errors:  make(map[string]int64),
		MinTime: time.Hour,
	}
	
	startTime := time.Now()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	
	// 启动并发测试
	for i := 0; i < st.config.Concurrency; i++ {
		wg.Add(1)
		go testFunc(&wg, stopCh)
	}
	
	// 等待测试时间结束
	time.Sleep(st.config.Duration)
	close(stopCh)
	wg.Wait()
	
	// 计算结果
	st.result.TotalTime = time.Since(startTime)
	if st.result.TotalRequests > 0 {
		st.result.QPS = float64(st.result.TotalRequests) / st.result.TotalTime.Seconds()
		st.result.SuccessRate = float64(st.result.SuccessRequests) / float64(st.result.TotalRequests) * 100
		st.result.AvgTime = time.Duration(int64(st.result.TotalTime) / st.result.TotalRequests)
	}
	
	// 输出结果
	st.printResult()
}

// printResult 打印测试结果
func (st *StressTest) printResult() {
	fmt.Printf("\n测试结果:\n")
	fmt.Printf("总请求数: %d\n", st.result.TotalRequests)
	fmt.Printf("成功请求数: %d\n", st.result.SuccessRequests)
	fmt.Printf("失败请求数: %d\n", st.result.FailedRequests)
	fmt.Printf("成功率: %.2f%%\n", st.result.SuccessRate)
	fmt.Printf("QPS: %.2f\n", st.result.QPS)
	fmt.Printf("平均响应时间: %v\n", st.result.AvgTime)
	fmt.Printf("最小响应时间: %v\n", st.result.MinTime)
	fmt.Printf("最大响应时间: %v\n", st.result.MaxTime)
	
	if len(st.result.Errors) > 0 {
		fmt.Printf("\n错误统计:\n")
		for err, count := range st.result.Errors {
			fmt.Printf("  %s: %d次\n", err, count)
		}
	}
	fmt.Println("----------------------------------------")
}

// runMixedTest 运行混合场景测试
func (st *StressTest) runMixedTest() {
	fmt.Printf("\n=== 混合场景压力测试 ===\n")
	fmt.Printf("并发数: %d, 持续时间: %v\n", st.config.Concurrency, st.config.Duration)
	
	// 重置结果
	st.result = &TestResult{
		Errors:  make(map[string]int64),
		MinTime: time.Hour,
	}
	
	startTime := time.Now()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	
	// 分配不同类型的测试
	healthWorkers := st.config.Concurrency / 4
	createWorkers := st.config.Concurrency / 4
	submitWorkers := st.config.Concurrency / 4
	queryWorkers := st.config.Concurrency - healthWorkers - createWorkers - submitWorkers
	
	// 启动健康检查测试
	for i := 0; i < healthWorkers; i++ {
		wg.Add(1)
		go st.testHealthCheck(&wg, stopCh)
	}
	
	// 启动创建排行榜测试
	for i := 0; i < createWorkers; i++ {
		wg.Add(1)
		go st.testCreateLeaderboard(&wg, stopCh)
	}
	
	// 启动提交分数测试
	for i := 0; i < submitWorkers; i++ {
		wg.Add(1)
		go st.testSubmitScore(&wg, stopCh)
	}
	
	// 启动查询排行榜测试
	for i := 0; i < queryWorkers; i++ {
		wg.Add(1)
		go st.testGetLeaderboard(&wg, stopCh)
	}
	
	// 等待测试时间结束
	time.Sleep(st.config.Duration)
	close(stopCh)
	wg.Wait()
	
	// 计算结果
	st.result.TotalTime = time.Since(startTime)
	if st.result.TotalRequests > 0 {
		st.result.QPS = float64(st.result.TotalRequests) / st.result.TotalTime.Seconds()
		st.result.SuccessRate = float64(st.result.SuccessRequests) / float64(st.result.TotalRequests) * 100
		st.result.AvgTime = time.Duration(int64(st.result.TotalTime) / st.result.TotalRequests)
	}
	
	// 输出结果
	st.printResult()
}

func main() {
	baseURL := "http://localhost:8080"
	
	fmt.Println("🚀 排行榜系统压力测试开始")
	fmt.Printf("目标服务器: %s\n", baseURL)
	
	// 等待服务器准备
	fmt.Println("\n等待服务器准备...")
	time.Sleep(2 * time.Second)
	
	// 测试不同并发级别
	concurrencyLevels := []int{10, 50, 100, 200}
	testDuration := 30 * time.Second
	
	for _, concurrency := range concurrencyLevels {
		fmt.Printf("\n\n🔥 开始 %d 并发测试\n", concurrency)
		fmt.Println("==========================================")
		
		config := &TestConfig{
			BaseURL:     baseURL,
			Concurrency: concurrency,
			Duration:    testDuration,
			Timeout:     10 * time.Second,
		}
		
		stressTest := NewStressTest(config)
		
		// 1. 健康检查测试
		stressTest.runTest("健康检查", stressTest.testHealthCheck)
		
		// 2. 创建排行榜测试
		stressTest.runTest("创建排行榜", stressTest.testCreateLeaderboard)
		
		// 3. 提交分数测试
		stressTest.runTest("提交分数", stressTest.testSubmitScore)
		
		// 4. 查询排行榜测试
		stressTest.runTest("查询排行榜", stressTest.testGetLeaderboard)
		
		// 5. 混合场景测试
		stressTest.runMixedTest()
		
		// 等待一段时间再进行下一轮测试
		if concurrency < concurrencyLevels[len(concurrencyLevels)-1] {
			fmt.Println("\n等待系统恢复...")
			time.Sleep(5 * time.Second)
		}
	}
	
	fmt.Println("\n\n🎉 压力测试完成！")
	fmt.Println("\n测试总结:")
	fmt.Println("- 测试了 4 种不同的并发级别 (10, 50, 100, 200)")
	fmt.Println("- 每个级别测试了 5 种场景 (健康检查、创建排行榜、提交分数、查询排行榜、混合场景)")
	fmt.Println("- 每个测试持续 30 秒")
	fmt.Println("- 记录了响应时间、成功率、QPS 等关键指标")
	fmt.Println("\n建议根据测试结果优化系统性能和调整服务器配置。")
}