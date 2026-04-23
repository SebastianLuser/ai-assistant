package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type ExternalHookDef struct {
	Name       string            `yaml:"name"`
	Event      string            `yaml:"event"`
	Type       string            `yaml:"type"` // "command", "webhook"
	Command    string            `yaml:"command"`
	URL        string            `yaml:"url"`
	Method     string            `yaml:"method"`
	Timeout    string            `yaml:"timeout"`
	Async      bool              `yaml:"async"`
	Conditions map[string]string `yaml:"conditions"`
}

func (d ExternalHookDef) timeout() time.Duration {
	if d.Timeout == "" {
		return 10 * time.Second
	}
	dur, err := time.ParseDuration(d.Timeout)
	if err != nil {
		return 10 * time.Second
	}
	return dur
}

func (d ExternalHookDef) method() string {
	if d.Method == "" {
		return http.MethodPost
	}
	return d.Method
}

func executeExternal(ctx context.Context, def ExternalHookDef, event Event) error {
	switch def.Type {
	case "command":
		return executeCommand(ctx, def, event)
	case "webhook":
		return executeWebhook(ctx, def, event)
	default:
		return fmt.Errorf("unknown hook type: %s", def.Type)
	}
}

func executeCommand(ctx context.Context, def ExternalHookDef, event Event) error {
	ctx, cancel := context.WithTimeout(ctx, def.timeout())
	defer cancel()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", def.Command)
	cmd.Stdin = bytes.NewReader(payload)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %w (output: %s)", def.Command, err, string(output))
	}
	return nil
}

func executeWebhook(ctx context.Context, def ExternalHookDef, event Event) error {
	ctx, cancel := context.WithTimeout(ctx, def.timeout())
	defer cancel()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, def.method(), def.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook %s failed: %w", def.URL, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook %s returned %d", def.URL, resp.StatusCode)
	}
	return nil
}

func matchConditions(conditions map[string]string, payload any) bool {
	if len(conditions) == 0 {
		return true
	}

	data, ok := payload.(map[string]any)
	if !ok {
		b, err := json.Marshal(payload)
		if err != nil {
			return false
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return false
		}
	}

	for key, expected := range conditions {
		val, exists := data[key]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", val) != expected {
			return false
		}
	}
	return true
}

func runExternalHook(ctx context.Context, def ExternalHookDef, event Event) {
	if !matchConditions(def.Conditions, event.Payload) {
		return
	}

	if def.Async {
		go func() {
			if err := executeExternal(ctx, def, event); err != nil {
				log.Printf("hooks: external %q error: %v", def.Name, err)
			}
		}()
		return
	}

	if err := executeExternal(ctx, def, event); err != nil {
		log.Printf("hooks: external %q error: %v", def.Name, err)
	}
}
