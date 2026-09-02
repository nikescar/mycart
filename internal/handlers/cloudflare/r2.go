package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	var result struct {
		Result struct {
			Buckets []R2Bucket `json:"buckets"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result.Buckets, nil
}

func (c *R2Client) CreateBucket(name string) (*R2Bucket, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result R2Bucket `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
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
