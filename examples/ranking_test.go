// Package examples
// 排行榜功能测试示例

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestClient 测试客户端
type TestClient struct {
	baseURL string
	client  *http.Client
}

// NewTestClient 创建测试客户端
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// makeRequest 发送HTTP请求
func (tc *TestClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, tc.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return tc.client.Do(req)
}

// testHealthCheck 测试健康检查
func (tc *TestClient) testHealthCheck() error {
	fmt.Println("\n=== 测试健康检查 ===")
	resp, err := tc.makeRequest("GET", "/health", nil)
	if err != nil {
		return fmt.Errorf("健康检查请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println("✓ 健康检查通过")
	return nil
}

// testCreateLeaderboard 测试创建排行榜
func (tc *TestClient) testCreateLeaderboard() (string, error) {
	fmt.Println("\n=== 测试创建排行榜 ===")
	leaderboard := map[string]interface{}{
		"name":        "测试排行榜",
		"description": "这是一个测试排行榜",
		"game_id":     "test_game_001",
		"type":        "score",
		"order":       "desc",
		"max_entries": 100,
	}

	resp, err := tc.makeRequest("POST", "/api/v1/leaderboard/create", leaderboard)
	if err != nil {
		return "", fmt.Errorf("创建排行榜请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("创建排行榜失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应获取排行榜ID
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(string); ok {
				fmt.Printf("✓ 排行榜创建成功，ID: %s\n", id)
				return id, nil
			}
		}
	}

	fmt.Println("✓ 排行榜创建成功")
	return "test_leaderboard_id", nil
}

// testSubmitScore 测试提交分数
func (tc *TestClient) testSubmitScore(leaderboardID string) error {
	fmt.Println("\n=== 测试提交分数 ===")
	scores := []map[string]interface{}{
		{
			"user_id":        "user_001",
			"username":       "玩家1",
			"score":          1000,
			"leaderboard_id": leaderboardID,
		},
		{
			"user_id":        "user_002",
			"username":       "玩家2",
			"score":          1500,
			"leaderboard_id": leaderboardID,
		},
		{
			"user_id":        "user_003",
			"username":       "玩家3",
			"score":          800,
			"leaderboard_id": leaderboardID,
		},
	}

	for i, score := range scores {
		resp, err := tc.makeRequest("POST", "/api/v1/score/submit", score)
		if err != nil {
			fmt.Printf("提交分数 %d 失败: %v\n", i+1, err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("分数 %d - 状态码: %d, 响应: %s\n", i+1, resp.StatusCode, string(body))
	}

	fmt.Println("✓ 分数提交完成")
	return nil
}

// testGetLeaderboard 测试获取排行榜
func (tc *TestClient) testGetLeaderboard() error {
	fmt.Println("\n=== 测试获取排行榜列表 ===")
	resp, err := tc.makeRequest("GET", "/api/v1/leaderboard/list", nil)
	if err != nil {
		return fmt.Errorf("获取排行榜列表失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))

	fmt.Println("✓ 排行榜列表获取完成")
	return nil
}

// testMetrics 测试指标接口
func (tc *TestClient) testMetrics() error {
	fmt.Println("\n=== 测试指标接口 ===")
	resp, err := tc.makeRequest("GET", "/api/v1/metrics", nil)
	if err != nil {
		return fmt.Errorf("获取指标失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))

	fmt.Println("✓ 指标接口测试完成")
	return nil
}

func main() {
	baseURL := "http://localhost:8080"
	client := NewTestClient(baseURL)

	fmt.Println("开始测试排行榜系统...")
	fmt.Printf("服务器地址: %s\n", baseURL)

	// 等待服务器启动
	fmt.Println("\n等待服务器启动...")
	time.Sleep(2 * time.Second)

	// 测试健康检查
	if err := client.testHealthCheck(); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}

	// 测试创建排行榜
	leaderboardID, err := client.testCreateLeaderboard()
	if err != nil {
		fmt.Printf("❌ 创建排行榜失败: %v\n", err)
		// 继续其他测试
	}

	// 测试提交分数
	if leaderboardID != "" {
		if err := client.testSubmitScore(leaderboardID); err != nil {
			fmt.Printf("❌ 提交分数失败: %v\n", err)
		}
	}

	// 测试获取排行榜
	if err := client.testGetLeaderboard(); err != nil {
		fmt.Printf("❌ 获取排行榜失败: %v\n", err)
	}

	// 测试指标接口
	if err := client.testMetrics(); err != nil {
		fmt.Printf("❌ 指标接口测试失败: %v\n", err)
	}

	fmt.Println("\n🎉 排行榜系统测试完成！")
}