package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ============================================================================
// SPLUNK API SCHEMAS
// ============================================================================
type SplunkSearchResponse struct {
	Results []map[string]interface{} `json:"results"`
}

// ============================================================================
// GITHUB MODELS API SCHEMAS
// ============================================================================
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ============================================================================
// CORE ANALYZER FUNCTION
// ============================================================================
func RunSplunkAIAnalyzer() error {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Warning: No .env file found, relying on system environment variables.")
	}

	splunkURL := os.Getenv("SPLUNK_URL")
	splunkUser := os.Getenv("SPLUNK_USER")
	splunkPass := os.Getenv("SPLUNK_PASSWORD")
	githubToken := os.Getenv("GITHUB_TOKEN")

	if splunkURL == "" || splunkUser == "" || splunkPass == "" || githubToken == "" {
		return fmt.Errorf("missing required environment variables (SPLUNK_URL, SPLUNK_USER, SPLUNK_PASSWORD, GITHUB_TOKEN)")
	}

	// 2. Configure HTTP Client (Allow self-signed certs for Splunk localhost)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}

	// 3. Execute Splunk Query (Oneshot mode for synchronous results)
	fmt.Println("🔍 Fetching telemetry from Splunk Enterprise...")
	
	searchQuery := `search source="simulation_telemetry.jsonl" | sort - time_sec | head 10`
	data := url.Values{}
	data.Set("search", searchQuery)
	data.Set("exec_mode", "oneshot")
	data.Set("output_mode", "json")

	reqURL := fmt.Sprintf("%s/services/search/jobs", strings.TrimSuffix(splunkURL, "/"))
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Splunk request: %w", err)
	}

	req.SetBasicAuth(splunkUser, splunkPass)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("splunk API request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Splunk response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("splunk returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var splunkResp SplunkSearchResponse
	if err := json.Unmarshal(bodyBytes, &splunkResp); err != nil {
		return fmt.Errorf("failed to decode Splunk JSON: %w", err)
	}

	if len(splunkResp.Results) == 0 {
		return fmt.Errorf("no results found in Splunk for the given query")
	}

	// 4. Extract & Format Telemetry payload
	var telemetryBuilder strings.Builder
	for i, row := range splunkResp.Results {
		telemetryBuilder.WriteString(fmt.Sprintf("Row %d:\n", i+1))
		telemetryBuilder.WriteString(fmt.Sprintf("- Event: %v\n", extractField(row, "event")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Risk Score: %v\n", extractField(row, "risk_score")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Physical Queue: %v\n", extractField(row, "physical_queue")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Physical Latency: %v\n", extractField(row, "physical_latency")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Retry Pool: %v\n", extractField(row, "retry_pool")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Pods Ready: %v\n", extractField(row, "pods_ready")))
		telemetryBuilder.WriteString(fmt.Sprintf("- Pods Target: %v\n\n", extractField(row, "pods_target")))
	}

	// 5. Build AI Prompt
	prompt := fmt.Sprintf(`You are a senior SRE incident analyst.

Analyze the operational telemetry from Splunk.

TELEMETRY DATA:
%s

Provide:
- Severity
- Root Cause
- Business Impact
- Recommended Actions

Keep response under 150 words.`, telemetryBuilder.String())

	// 6. Send to GitHub Models API (DeepSeek-V3)
	fmt.Println("🧠 Forwarding telemetry to DeepSeek-V3 for Root Cause Analysis...")
	
	chatReq := ChatRequest{
		Model: "DeepSeek-V3", 
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.1, // Low temperature for factual analysis
		MaxTokens:   300,
	}

	jsonPayload, err := json.Marshal(chatReq)
	if err != nil {
		return fmt.Errorf("failed to marshal GitHub API request: %w", err)
	}

	ghReq, err := http.NewRequest("POST", "https://models.inference.ai.azure.com/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub API request: %w", err)
	}

	ghReq.Header.Set("Content-Type", "application/json")
	ghReq.Header.Set("Authorization", "Bearer "+githubToken)

	ghResp, err := client.Do(ghReq)
	if err != nil {
		return fmt.Errorf("github API request failed: %w", err)
	}
	defer ghResp.Body.Close()

	ghBodyBytes, err := io.ReadAll(ghResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read GitHub API response: %w", err)
	}

	if ghResp.StatusCode != http.StatusOK {
		return fmt.Errorf("github API returned status %d: %s", ghResp.StatusCode, string(ghBodyBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(ghBodyBytes, &chatResp); err != nil {
		return fmt.Errorf("failed to decode GitHub API JSON: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("no response generated by DeepSeek-V3")
	}

	// 7. Print Output
	fmt.Println("\n===============================================================================")
	fmt.Println(" 🤖 SRE INCIDENT ANALYSIS (DeepSeek-V3 via Splunk Telemetry)")
	fmt.Println("===============================================================================")
	fmt.Println(chatResp.Choices[0].Message.Content)
	fmt.Println("===============================================================================")

	return nil
}

// Helper function to safely extract Splunk fields, which can sometimes be returned as arrays of strings
func extractField(row map[string]interface{}, key string) string {
	val, exists := row[key]
	if !exists {
		return "N/A"
	}
	
	switch v := val.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			return fmt.Sprintf("%v", v[0])
		}
	}
	return fmt.Sprintf("%v", val)
}