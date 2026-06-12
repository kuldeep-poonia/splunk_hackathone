package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	token := os.Getenv("GITHUB_TOKEN")

	body := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Say hello in one sentence.",
			},
		},
		"model": "gpt-4o-mini",
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest(
		"POST",
		"https://models.inference.ai.azure.com/chat/completions",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	fmt.Println("Status:", resp.Status)
	fmt.Println(string(data))
}