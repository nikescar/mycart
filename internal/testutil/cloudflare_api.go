package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

type CloudflareClient struct {
	accountID  string
	apiToken   string
	httpClient *http.Client
}

func NewCloudflareClient(accountID, apiToken string) *CloudflareClient {
	return &CloudflareClient{
		accountID:  accountID,
		apiToken:   apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateD1Database creates a new D1 database and returns its ID
func (c *CloudflareClient) CreateD1Database(name string) (string, error) {
	url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)

	body := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create D1 database request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create D1 database failed (HTTP %d): %s", resp.StatusCode, respBody)
	}

	var result struct {
		Result struct {
			UUID string `json:"uuid"`
			Name string `json:"name"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse D1 create response: %w", err)
	}

	return result.Result.UUID, nil
}

// DeleteD1Database deletes a D1 database (idempotent - 404 is OK)
func (c *CloudflareClient) DeleteD1Database(databaseID string) error {
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s", cloudflareAPIBase, c.accountID, databaseID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete D1 database request failed: %w", err)
	}
	defer resp.Body.Close()

	// 404 is OK - resource already deleted (idempotent)
	if resp.StatusCode == 404 {
		return nil
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete D1 database failed (HTTP %d): %s", resp.StatusCode, respBody)
	}

	return nil
}

// CreateR2Bucket creates a new R2 bucket
func (c *CloudflareClient) CreateR2Bucket(name string) error {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)

	body := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create R2 bucket request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create R2 bucket failed (HTTP %d): %s", resp.StatusCode, respBody)
	}

	return nil
}

// DeleteR2Bucket deletes an R2 bucket (idempotent - 404 is OK)
func (c *CloudflareClient) DeleteR2Bucket(name string) error {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s", cloudflareAPIBase, c.accountID, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete R2 bucket request failed: %w", err)
	}
	defer resp.Body.Close()

	// 404 is OK - resource already deleted (idempotent)
	if resp.StatusCode == 404 {
		return nil
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete R2 bucket failed (HTTP %d): %s", resp.StatusCode, respBody)
	}

	return nil
}

// ExecuteD1SQL executes SQL statements on a D1 database
func (c *CloudflareClient) ExecuteD1SQL(databaseID, sql string) error {
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", cloudflareAPIBase, c.accountID, databaseID)

	body := map[string]string{"sql": sql}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute SQL request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("execute SQL failed (HTTP %d): %s", resp.StatusCode, respBody)
	}

	return nil
}
