package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	targetURL := "http://localhost:8080/api/v1/news" // 一个用于测试的URL，它会返回请求信息
	numConcurrentRequests := 100000         // 要并发发送的请求数量

	// 创建一个共享的 http.Client
	client := &http.Client{
		Timeout: 5 * time.Second, // 为每个请求设置一个超时
		Transport: &http.Transport{
			MaxIdleConns:          numConcurrentRequests, // 允许足够多的空闲连接
			IdleConnTimeout:       90 * time.Second,
			MaxConnsPerHost:       numConcurrentRequests, // 允许每个host有足够多的连接
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex // 用于保护 successCount 和 errorCount

	fmt.Printf("Starting %d concurrent requests to %s...\n", numConcurrentRequests, targetURL)

	startTime := time.Now()

	for i := 0; i < numConcurrentRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			// 使用 client.Get 发送请求
			resp, err := client.Get(targetURL)
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
				// fmt.Printf("Request %d failed: %v\n", requestID, err) // 可选：打印错误信息
				return
			}
			defer resp.Body.Close() // 务必关闭响应体

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				mu.Lock()
				successCount++
				mu.Unlock()
				// fmt.Printf("Request %d successful, Status: %s\n", requestID, resp.Status) // 可选：打印成功信息
			} else {
				mu.Lock()
				errorCount++
				mu.Unlock()
				// fmt.Printf("Request %d failed with status: %s\n", requestID, resp.Status) // 可选：打印状态码错误
			}
		}(i) // 传递请求ID以区分Goroutine
	}

	wg.Wait() // 等待所有请求完成
	duration := time.Since(startTime)

	fmt.Printf("\n--- Results ---\n")
	fmt.Printf("Total requests: %d\n", numConcurrentRequests)
	fmt.Printf("Successful requests: %d\n", successCount)
	fmt.Printf("Failed requests: %d\n", errorCount)
	fmt.Printf("Total duration: %s\n", duration)
	if numConcurrentRequests > 0 {
		fmt.Printf("Average time per request: %s\n", duration/time.Duration(numConcurrentRequests))
		fmt.Printf("Throughput: %.2f requests/second\n", float64(numConcurrentRequests)/duration.Seconds())
	}
}
