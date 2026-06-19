package llm

import "testing"

func TestAvailableLocalPrivateNetwork(t *testing.T) {
	c := NewClient(Config{
		Enabled: true,
		BaseURL: "https://192.168.0.64:11343/v1/chat/completions",
		Model:   "llama3.2:latest",
	})
	if !c.Available() {
		t.Fatal("expected local private-network LLM to be available without API key")
	}
	if c.cfg.BaseURL != "https://192.168.0.64:11343/v1" {
		t.Fatalf("base URL not normalized: %q", c.cfg.BaseURL)
	}
}

func TestAvailableCloudRequiresKey(t *testing.T) {
	c := NewClient(Config{
		Enabled: true,
		BaseURL: "https://api.openai.com/v1",
	})
	if c.Available() {
		t.Fatal("expected cloud LLM without API key to be unavailable")
	}
}
