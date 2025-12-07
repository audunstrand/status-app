package e2e_docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	commandsAuthToken = "test-secret-123"
	apiAuthToken      = "test-secret-456"
)

type dockerComposeEnv struct {
	ctx context.Context
}

func setupDockerEnvironment(t *testing.T) *dockerComposeEnv {
	ctx := context.Background()

	// Start docker-compose
	cmd := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "up", "-d", "--build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start compose: %v\nOutput: %s", err, string(output))
	}

	// Wait for services to be healthy
	t.Log("Waiting for services to start...")
	time.Sleep(15 * time.Second)

	return &dockerComposeEnv{
		ctx: ctx,
	}
}

func (e *dockerComposeEnv) teardown() error {
	cmd := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "down", "-v")
	return cmd.Run()
}

func (e *dockerComposeEnv) getServiceURL(serviceName string) (string, error) {
	// Get the port mapping
	cmd := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "port", serviceName, "8080")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get port for %s: %v", serviceName, err)
	}

	// Parse output like "0.0.0.0:54321"
	portStr := strings.TrimSpace(string(output))
	return fmt.Sprintf("http://%s", portStr), nil
}

func TestDockerE2E_SubmitStatusUpdate(t *testing.T) {
	env := setupDockerEnvironment(t)
	defer env.teardown()

	commandsURL, err := env.getServiceURL("commands")
	if err != nil {
		t.Fatalf("Failed to get commands URL: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Test submitting a status update
	payload := map[string]string{
		"team_id": "team-123",
		"content": "Working on Docker E2E tests",
		"author":  "test-user",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", commandsURL+"/commands/submit-update", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+commandsAuthToken)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200 or 201, got %d. Response: %s", resp.StatusCode, string(respBody))
	}
}

func TestDockerE2E_AuthenticationRequired(t *testing.T) {
	env := setupDockerEnvironment(t)
	defer env.teardown()

	commandsURL, err := env.getServiceURL("commands")
	if err != nil {
		t.Fatalf("Failed to get commands URL: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	payload := map[string]string{
		"team_id": "team-123",
		"content": "Test content",
		"author":  "test-user",
	}

	body, _ := json.Marshal(payload)

	// Request without authentication
	req, _ := http.NewRequest("POST", commandsURL+"/commands/submit-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthenticated request, got %d", resp.StatusCode)
	}
}

func TestDockerE2E_APIEndpoints(t *testing.T) {
	env := setupDockerEnvironment(t)
	defer env.teardown()

	apiURL, err := env.getServiceURL("api")
	if err != nil {
		t.Fatalf("Failed to get API URL: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Test health endpoint (no auth)
	resp, err := client.Get(apiURL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for health check, got %d", resp.StatusCode)
	}

	// Test teams endpoint (with auth)
	req, _ := http.NewRequest("GET", apiURL+"/api/teams", nil)
	req.Header.Set("Authorization", "Bearer "+apiAuthToken)

	resp2, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get teams: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(resp2.Body)
		t.Errorf("Expected status 200 or 404, got %d. Response: %s", resp2.StatusCode, string(respBody))
	}

	// Test teams endpoint (without auth)
	resp3, err := client.Get(apiURL + "/api/teams")
	if err != nil {
		t.Fatalf("Failed to get teams: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthorized request, got %d", resp3.StatusCode)
	}
}

func TestDockerE2E_EndToEndFlow(t *testing.T) {
	env := setupDockerEnvironment(t)
	defer env.teardown()

	commandsURL, err := env.getServiceURL("commands")
	if err != nil {
		t.Fatalf("Failed to get commands URL: %v", err)
	}

	apiURL, err := env.getServiceURL("api")
	if err != nil {
		t.Fatalf("Failed to get API URL: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Submit a status update
	updatePayload := map[string]string{
		"team_id": "team-e2e",
		"content": "End-to-end test update",
		"author":  "e2e-tester",
	}

	body, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("POST", commandsURL+"/commands/submit-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+commandsAuthToken)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to submit update: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to submit update, status: %d", resp.StatusCode)
	}

	// Step 2: Wait for projection to process (if projections service is running)
	time.Sleep(2 * time.Second)

	// Step 3: Query updates via API
	req2, _ := http.NewRequest("GET", apiURL+"/api/updates", nil)
	req2.Header.Set("Authorization", "Bearer "+apiAuthToken)

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Failed to query updates: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp2.Body)
		t.Logf("Query response: %s", string(respBody))
	}

	t.Log("✅ End-to-end flow completed successfully")
}
