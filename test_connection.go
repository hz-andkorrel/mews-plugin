package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func TestMewsConnection() {
	clientToken := os.Getenv("MEWS_CLIENT_TOKEN")
	accessToken := os.Getenv("MEWS_ACCESS_TOKEN")

	payload, _ := json.Marshal(map[string]string{
		"clientToken": clientToken,
		"accessToken": accessToken,
	})

	resp, err := http.Post("https://api.mews-demo.com/api/connector/v1/configuration/get", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, string(body))
}
