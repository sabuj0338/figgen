// Package telemetry records per-call LLM token usage so the cost impact of
// prompt/pruning optimizations can be measured over time. Records are appended
// to <outDir>/.figgen/usage.json.
package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Record is a single LLM call's usage entry.
type Record struct {
	Time         string `json:"time"`
	Stage        string `json:"stage"` // "plan" or "code"
	Label        string `json:"label"` // component/page name or chunk id
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	CachedTokens int    `json:"cached_tokens"`
}

var (
	mu      sync.Mutex
	outDir  string
	enabled bool
)

// Init enables telemetry and points it at the project output directory.
func Init(dir string) {
	mu.Lock()
	defer mu.Unlock()
	outDir = dir
	enabled = true
}

// Record appends a usage entry. It is a no-op until Init has been called and is
// safe to call concurrently.
func Add(r Record) {
	mu.Lock()
	defer mu.Unlock()
	if !enabled || outDir == "" {
		return
	}

	r.Time = time.Now().Format(time.RFC3339)
	path := filepath.Join(outDir, ".figgen", "usage.json")

	var records []Record
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &records)
	}
	records = append(records, r)

	if data, err := json.MarshalIndent(records, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(path), 0755)
		_ = os.WriteFile(path, data, 0644)
	}
}

// Summary aggregates usage from <outDir>/.figgen/usage.json.
type Summary struct {
	Calls        int
	InputTokens  int
	OutputTokens int
	CachedTokens int
	ByStage      map[string]int // stage -> total tokens
}

// Load reads and aggregates recorded usage for the given output directory.
func Load(dir string) (*Summary, []Record, error) {
	path := filepath.Join(dir, ".figgen", "usage.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, nil, err
	}

	s := &Summary{ByStage: map[string]int{}}
	for _, r := range records {
		s.Calls++
		s.InputTokens += r.InputTokens
		s.OutputTokens += r.OutputTokens
		s.CachedTokens += r.CachedTokens
		s.ByStage[r.Stage] += r.InputTokens + r.OutputTokens
	}

	return s, records, nil
}
