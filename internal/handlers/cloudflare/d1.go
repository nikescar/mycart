package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

type D1Client struct {
	accountID  string
	apiToken   string
	httpClient *http.Client
}

type D1Database struct {
	ID        string    `json:"uuid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewD1Client(accountID, apiToken string) *D1Client {
	return &D1Client{
		accountID:  accountID,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *D1Client) ListDatabases() ([]D1Database, error) {
	url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", body)
	}

	var result struct {
		Result []D1Database `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}

func (c *D1Client) CreateDatabase(name string) (*D1Database, error) {
	url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result D1Database `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return &result.Result, nil
}

func (c *D1Client) ExportDatabase(databaseID, outputPath string) error {
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/export", cloudflareAPIBase, c.accountID, databaseID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, _ := os.Create(outputPath)
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func (c *D1Client) ImportSQLite(databaseID, sqlitePath string) error {
	sqlData, _ := os.ReadFile(sqlitePath)
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", cloudflareAPIBase, c.accountID, databaseID)
	payload, _ := json.Marshal(map[string]string{"sql": string(sqlData)})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
