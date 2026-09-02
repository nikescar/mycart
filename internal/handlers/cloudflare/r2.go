package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type R2Client struct {
	accountID  string
	apiToken   string
	httpClient *http.Client
}

type R2Bucket struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"creation_date"`
}

type R2Object struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

func NewR2Client(accountID, apiToken string) *R2Client {
	return &R2Client{
		accountID:  accountID,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *R2Client) ListBuckets() ([]R2Bucket, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, body)
	}

	// Try parsing as array first (newer API format)
	var arrayResult struct {
		Success bool       `json:"success"`
		Result  []R2Bucket `json:"result"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &arrayResult); err == nil && arrayResult.Result != nil {
		return arrayResult.Result, nil
	}

	// Fall back to object format (older API format)
	var objectResult struct {
		Success bool `json:"success"`
		Result  struct {
			Buckets []R2Bucket `json:"buckets"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &objectResult); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return objectResult.Result.Buckets, nil
}

func (c *R2Client) CreateBucket(name string) (*R2Bucket, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
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
		Success bool        `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result R2Bucket `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	// Check Cloudflare API success field
	if !result.Success {
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", result.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
	}

	// Also check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Cloudflare R2 create bucket may return empty result, so we construct the bucket object
	// The frontend will reload the list anyway to get the full details
	if result.Result.Name == "" {
		return &R2Bucket{
			Name:      name,
			CreatedAt: time.Now(),
		}, nil
	}

	return &result.Result, nil
}

func (c *R2Client) ListObjects(bucketName string) ([]R2Object, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s/objects", cloudflareAPIBase, c.accountID, bucketName)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []R2Object `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}
