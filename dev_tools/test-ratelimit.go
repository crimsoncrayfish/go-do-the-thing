package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type RateLimitTester struct {
	baseURL string
	client  *http.Client
}

func NewRateLimitTester(baseURL string) *RateLimitTester {
	jar, _ := cookiejar.New(nil)
	return &RateLimitTester{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
	}
}

func (rt *RateLimitTester) Login(username, password string) error {
	loginURL := rt.baseURL + "/login"

	// Get login page first
	resp, err := rt.client.Get(loginURL)
	if err != nil {
		return fmt.Errorf("failed to get login page: %w", err)
	}
	resp.Body.Close()

	// Attempt login
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	resp, err = rt.client.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("failed to post login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body)[:200])
	}

	fmt.Printf("Login successful! Status: %d\n", resp.StatusCode)
	return nil
}

type RequestResult struct {
	ID           int
	StatusCode   int
	ResponseTime time.Duration
	Timestamp    time.Time
	Error        error
	RateLimited  bool
}

func (rt *RateLimitTester) makeRequest(endpoint string, id int) RequestResult {
	start := time.Now()
	resp, err := rt.client.Get(rt.baseURL + endpoint)
	duration := time.Since(start)

	result := RequestResult{
		ID:           id,
		ResponseTime: duration,
		Timestamp:    start,
		Error:        err,
	}

	if err != nil {
		return result
	}

	result.StatusCode = resp.StatusCode
	result.RateLimited = resp.StatusCode == 429

	// Read and close body
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return result
}

func (rt *RateLimitTester) SequentialTest(numRequests int, delay time.Duration, endpoint string) []RequestResult {
	fmt.Printf("\n=== Sequential Test: %d requests to %s ===\n", numRequests, endpoint)
	results := make([]RequestResult, 0, numRequests)

	for i := 0; i < numRequests; i++ {
		fmt.Printf("Sending request %d/%d...\n", i+1, numRequests)
		result := rt.makeRequest(endpoint, i+1)
		results = append(results, result)

		if result.Error != nil {
			fmt.Printf("  Error: %v\n", result.Error)
		} else {
			rateLimitedStr := ""
			if result.RateLimited {
				rateLimitedStr = " [RATE LIMITED]"
			}
			fmt.Printf("  Response: %d (%v)%s\n", result.StatusCode, result.ResponseTime, rateLimitedStr)
		}

		if i < numRequests-1 {
			time.Sleep(delay)
		}
	}

	return results
}

func (rt *RateLimitTester) ConcurrentTest(numRequests, numWorkers int, endpoint string) []RequestResult {
	fmt.Printf("\n=== Concurrent Test: %d requests with %d workers to %s ===\n", numRequests, numWorkers, endpoint)

	results := make([]RequestResult, numRequests)
	var wg sync.WaitGroup
	requestChan := make(chan int, numRequests)

	// Fill request channel
	for i := 0; i < numRequests; i++ {
		requestChan <- i + 1
	}
	close(requestChan)

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range requestChan {
				result := rt.makeRequest(endpoint, id)
				results[id-1] = result

				if result.Error != nil {
					fmt.Printf("  Request %d: ERROR - %v\n", result.ID, result.Error)
				} else {
					rateLimitedStr := ""
					if result.RateLimited {
						rateLimitedStr = " [RATE LIMITED]"
					}
					fmt.Printf("  Request %d: %d (%v)%s\n", result.ID, result.StatusCode, result.ResponseTime, rateLimitedStr)
				}
			}
		}()
	}

	wg.Wait()
	return results
}

func (rt *RateLimitTester) BurstTest(burstSize int, endpoint string) []RequestResult {
	fmt.Printf("\n=== Burst Test: %d requests as fast as possible to %s ===\n", burstSize, endpoint)

	results := make([]RequestResult, burstSize)
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result := rt.makeRequest(endpoint, id+1)
			results[id] = result
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	rateLimitedCount := 0
	errorCount := 0
	successCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
			if result.RateLimited {
				rateLimitedCount++
			}
		}
	}

	fmt.Printf("Burst completed in %v\n", duration)
	fmt.Printf("Results: %d successful, %d rate limited, %d errors\n", successCount, rateLimitedCount, errorCount)

	return results
}

func printSummary(results []RequestResult, testName string) {
	if len(results) == 0 {
		return
	}

	fmt.Printf("\n=== %s Summary ===\n", testName)

	statusCodes := make(map[int]int)
	var responseTimes []time.Duration
	rateLimited := 0
	errors := 0

	for _, result := range results {
		if result.Error != nil {
			errors++
		} else {
			statusCodes[result.StatusCode]++
			responseTimes = append(responseTimes, result.ResponseTime)
			if result.RateLimited {
				rateLimited++
			}
		}
	}

	fmt.Printf("Total requests: %d\n", len(results))
	fmt.Printf("Status codes: %v\n", statusCodes)
	fmt.Printf("Rate limited responses: %d\n", rateLimited)
	fmt.Printf("Errors: %d\n", errors)

	if len(responseTimes) > 0 {
		var total time.Duration
		min := responseTimes[0]
		max := responseTimes[0]

		for _, rt := range responseTimes {
			total += rt
			if rt < min {
				min = rt
			}
			if rt > max {
				max = rt
			}
		}

		avg := total / time.Duration(len(responseTimes))
		fmt.Printf("Response times: avg=%v, min=%v, max=%v\n", avg, min, max)
	}
}

func runRateLimitTest() {
	var (
		baseURL   = flag.String("url", "http://localhost:8080", "Base URL of the application")
		username  = flag.String("username", "", "Username for login (required)")
		password  = flag.String("password", "", "Password for login (required)")
		endpoint  = flag.String("endpoint", "/", "Endpoint to test")
		skipLogin = flag.Bool("skip-login", false, "Skip login step")
	)
	flag.Parse()

	if !*skipLogin && (*username == "" || *password == "") {
		fmt.Println("Username and password are required unless --skip-login is used")
		flag.Usage()
		os.Exit(1)
	}

	tester := NewRateLimitTester(*baseURL)

	// Login first (unless skipped)
	if !*skipLogin {
		fmt.Println("Attempting to login...")
		if err := tester.Login(*username, *password); err != nil {
			fmt.Printf("Login failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Skipping login step...")
	}

	fmt.Printf("Testing rate limiting on %s%s\n", *baseURL, *endpoint)

	// Test 1: Sequential requests
	seqResults := tester.SequentialTest(10, 200*time.Millisecond, *endpoint)
	printSummary(seqResults, "Sequential Test")

	// Test 2: Concurrent requests
	concResults := tester.ConcurrentTest(20, 5, *endpoint)
	printSummary(concResults, "Concurrent Test")

	// Test 3: Burst test
	burstResults := tester.BurstTest(50, *endpoint)
	printSummary(burstResults, "Burst Test")

	fmt.Println("\n=== Rate Limit Testing Complete ===")
}
