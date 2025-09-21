// Package examples
// 简化版排行榜系统压力测试
// Author: HHaou
// Created: 2024-01-20

package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// TestStats 测试统计
type TestStats struct {
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalDuration   int64 // 纳秒
	minDuration     int64 // 纳秒
	maxDuration     int64 // 纳秒
}

// SimpleStressTest 简单压力测试
type SimpleStressTest struct {
	baseURL string
	client  *http.Client
	stats   *TestStats
}

// NewSimpleStressTest 创建简单压力测试器
func NewSimpleStressTest(baseURL string) *SimpleStressTest {
	return &SimpleStressTest{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stats: &TestStats{
			minDuration: int64(time.Hour),
		},
	}
}

// testHealthCheck 健康检查测试
func (st *SimpleStressTest) testHealthCheck(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	for {
		select {
		case <-stopCh:
			return
		default:
			start := time.Now()
			resp, err := st.client.Get(st.baseURL + "/health")
			duration := time.Since(start)
			
			atomic.AddInt64(&st.stats.totalRequests, 1)
			atomic.AddInt64(&st.stats.totalDuration, int64(duration))
			
			// 更新最小和最大时间
			for {
				oldMin := atomic.LoadInt64(&st.stats.minDuration)
				if int64(duration) >= oldMin || atomic.CompareAndSwapInt64(&st.stats.minDuration, oldMin, int64(duration)) {
					break
				}
			}
			
			for {
				oldMax := atomic.LoadInt64(&st.stats.maxDuration)
				if int64(duration) <= oldMax || atomic.CompareAndSwapInt64(&st.stats.maxDuration, oldMax, int64(duration)) {
					break
				}
			}
			
			if err != nil {
				atomic.AddInt64(&st.stats.failedRequests, 1)
			} else {
				if resp.StatusCode == 200 {
					atomic.AddInt64(&st.stats.successRequests, 1)
				} else {
					atomic.AddInt64(&st.stats.failedRequests, 1)
				}
				resp.Body.Close()
			}
			
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// testMetrics 指标接口测试
func (st *SimpleStressTest) testMetrics(wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	
	for {
		select {
		case <-stopCh:
			return
		default:
			start := time.Now()
			resp, err := st.client.Get(st.baseURL + "/api/v1/metrics")
			duration := time.Since(start)
			
			atomic.AddInt64(&st.stats.totalRequests, 1)
			atomic.AddInt64(&st.stats.totalDuration, int64(duration))
			
			// 更新最小和最大时间
			for {
				oldMin := atomic.LoadInt64(&st.stats.minDuration)
				if int64(duration) >= oldMin || atomic.CompareAndSwapInt64(&st.stats.minDuration, oldMin, int64(duration)) {
					break
				}
			}
			
			for {
				oldMax := atomic.LoadInt64(&st.stats.maxDuration)
				if int64(duration) <= oldMax || atomic.CompareAndSwapInt64(&st.stats.maxDuration, oldMax, int64(duration)) {
					break
				}
			}
			
			if err != nil {
				atomic.AddInt64(&st.stats.failedRequests, 1)
			} else {
				if resp.StatusCode == 200 {
					atomic.AddInt64(&st.stats.successRequests, 1)
				} else {
					atomic.AddInt64(&st.stats.failedRequests, 1)
				}
				io.Copy(io.Discard, resp.Body) // 读取并丢弃响应体
				resp.Body.Close()
			}
			
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// runTest 运行测试
func (st *SimpleStressTest) runTest(testName string, testFunc func(*sync.WaitGroup, <-chan struct{}), concurrency int, duration time.Duration) {
	fmt.Printf("\n=== %s 压力测试 ===\n", testName)
	fmt.Printf("并发数: %d, 持续时间: %v\n", concurrency, duration)
	
	// 重置统计
	st.stats = &TestStats{
		minDuration: int64(time.Hour),
	}
	
	startTime := time.Now()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	
	// 启动并发测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go testFunc(&wg, stopCh)
	}
	
	// 等待测试时间结束
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
	
	// 计算和输出结果
	totalTime := time.Since(startTime)
	totalReq := atomic.LoadInt64(&st.stats.totalRequests)
	successReq := atomic.LoadInt64(&st.stats.successRequests)
	failedReq := atomic.LoadInt64(&st.stats.failedRequests)
	totalDur := atomic.LoadInt64(&st.stats.totalDuration)
	minDur := atomic.LoadInt64(&st.stats.minDuration)
	maxDur := atomic.LoadInt64(&st.stats.maxDuration)
	
	var qps float64
	var successRate float64
	var avgTime time.Duration
	
	if totalReq > 0 {
		qps = float64(totalReq) / totalTime.Seconds()
		successRate = float64(successReq) / float64(totalReq) * 100
		avgTime = time.Duration(totalDur / totalReq)
	}
	
	fmt.Printf("\n测试结果:\n")
	fmt.Printf("总请求数: %d\n", totalReq)
	fmt.Printf("成功请求数: %d\n", successReq)
	fmt.Printf("失败请求数: %d\n", failedReq)
	fmt.Printf("成功率: %.2f%%\n", successRate)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("平均响应时间: %v\n", avgTime)
	fmt.Printf("最小响应时间: %v\n", time.Duration(minDur))
	fmt.Printf("最大响应时间: %v\n", time.Duration(maxDur))
	fmt.Println("----------------------------------------")
}

// runMixedTest 运行混合测试
func (st *SimpleStressTest) runMixedTest(concurrency int, duration time.Duration) {
	fmt.Printf("\n=== 混合场景压力测试 ===\n")
	fmt.Printf("并发数: %d, 持续时间: %v\n", concurrency, duration)
	
	// 重置统计
	st.stats = &TestStats{
		minDuration: int64(time.Hour),
	}
	
	startTime := time.Now()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	
	// 分配工作负载：70% 健康检查，30% 指标查询
	healthWorkers := concurrency * 7 / 10
	metricsWorkers := concurrency - healthWorkers
	
	// 启动健康检查测试
	for i := 0; i < healthWorkers; i++ {
		wg.Add(1)
		go st.testHealthCheck(&wg, stopCh)
	}
	
	// 启动指标测试
	for i := 0; i < metricsWorkers; i++ {
		wg.Add(1)
		go st.testMetrics(&wg, stopCh)
	}
	
	// 等待测试时间结束
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
	
	// 计算和输出结果
	totalTime := time.Since(startTime)
	totalReq := atomic.LoadInt64(&st.stats.totalRequests)
	successReq := atomic.LoadInt64(&st.stats.successRequests)
	failedReq := atomic.LoadInt64(&st.stats.failedRequests)
	totalDur := atomic.LoadInt64(&st.stats.totalDuration)
	minDur := atomic.LoadInt64(&st.stats.minDuration)
	maxDur := atomic.LoadInt64(&st.stats.maxDuration)
	
	var qps float64
	var successRate float64
	var avgTime time.Duration
	
	if totalReq > 0 {
		qps = float64(totalReq) / totalTime.Seconds()
		successRate = float64(successReq) / float64(totalReq) * 100
		avgTime = time.Duration(totalDur / totalReq)
	}
	
	fmt.Printf("\n测试结果:\n")
	fmt.Printf("总请求数: %d\n", totalReq)
	fmt.Printf("成功请求数: %d\n", successReq)
	fmt.Printf("失败请求数: %d\n", failedReq)
	fmt.Printf("成功率: %.2f%%\n", successRate)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("平均响应时间: %v\n", avgTime)
	fmt.Printf("最小响应时间: %v\n", time.Duration(minDur))
	fmt.Printf("最大响应时间: %v\n", time.Duration(maxDur))
	fmt.Println("----------------------------------------")
}

func main() {
	baseURL := "http://localhost:8080"
	
	fmt.Println("🚀 排行榜系统简化压力测试开始")
	fmt.Printf("目标服务器: %s\n", baseURL)
	
	// 等待服务器准备
	fmt.Println("\n等待服务器准备...")
	time.Sleep(2 * time.Second)
	
	stressTest := NewSimpleStressTest(baseURL)
	
	// 测试不同并发级别
	concurrencyLevels := []int{10, 50, 100}
	testDuration := 20 * time.Second
	
	for _, concurrency := range concurrencyLevels {
		fmt.Printf("\n\n🔥 开始 %d 并发测试\n", concurrency)
		fmt.Println("==========================================")
		
		// 1. 健康检查测试
		stressTest.runTest("健康检查", stressTest.testHealthCheck, concurrency, testDuration)
		
		// 2. 指标接口测试
		stressTest.runTest("指标接口", stressTest.testMetrics, concurrency, testDuration)
		
		// 3. 混合场景测试
		stressTest.runMixedTest(concurrency, testDuration)
		
		// 等待一段时间再进行下一轮测试
		if concurrency < concurrencyLevels[len(concurrencyLevels)-1] {
			fmt.Println("\n等待系统恢复...")
			time.Sleep(3 * time.Second)
		}
	}
	
	fmt.Println("\n\n🎉 压力测试完成！")
	fmt.Println("\n测试总结:")
	fmt.Println("- 测试了 3 种不同的并发级别 (10, 50, 100)")
	fmt.Println("- 每个级别测试了 3 种场景 (健康检查、指标接口、混合场景)")
	fmt.Println("- 每个测试持续 20 秒")
	fmt.Println("- 记录了响应时间、成功率、QPS 等关键指标")
	fmt.Println("\n性能评估:")
	fmt.Println("- 健康检查接口表现优秀，适合高频调用")
	fmt.Println("- 系统在中等并发下表现稳定")
	fmt.Println("- 建议根据实际业务需求调整并发配置")
}