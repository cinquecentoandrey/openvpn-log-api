package vpn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Client struct {
	LogdbaPath string
	LogDBPath  string
	Timeout    time.Duration
}

func NewClient(logdbaPath string, logDBPath string, timeoutSec int) *Client {
	return &Client{
		LogdbaPath: logdbaPath,
		LogDBPath:  logDBPath,
		Timeout:    time.Duration(timeoutSec) * time.Second,
	}
}

func (c *Client) GetLogs(dateFrom, dateTo string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	var args []string
	if c.LogDBPath != "" {
		args = append(args, fmt.Sprintf("--db=%s", c.LogDBPath))
	}

	args = append(args, "--json",
		"--start_time_ge", dateFrom,
		"--start_time_lt", dateTo,
	)

	cmd := exec.CommandContext(ctx, c.LogdbaPath, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("logdba error: %v (%s)", err, stderr.String())
	}

	var raw [][]any
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %v", err)
	}

	if len(raw) < 2 {
		return []map[string]any{}, nil
	}

	headers := make([]string, len(raw[0]))
	for i, h := range raw[0] {
		headers[i] = fmt.Sprintf("%v", h)
	}

	var logs []map[string]any
	for _, row := range raw[1:] {
		entry := make(map[string]any)
		for i, val := range row {
			if i < len(headers) {
				entry[headers[i]] = val
			}
		}
		logs = append(logs, entry)
	}

	return logs, nil
}
