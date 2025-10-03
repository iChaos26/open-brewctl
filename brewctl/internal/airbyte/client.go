package airbyte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AirbyteClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ConnectionRequest struct {
	Name          string       `json:"name"`
	SourceID      string       `json:"sourceId"`
	DestinationID string       `json:"destinationId"`
	SyncCatalog   SyncCatalog  `json:"syncCatalog"`
	ScheduleType  string       `json:"scheduleType"`
	ScheduleData  ScheduleData `json:"scheduleData"`
	Status        string       `json:"status"`
}

type SyncCatalog struct {
	Streams []StreamConfig `json:"streams"`
}

type StreamConfig struct {
	Stream Stream             `json:"stream"`
	Config StreamConfigDetail `json:"config"`
}

type Stream struct {
	Name                    string                 `json:"name"`
	JSONSchema              map[string]interface{} `json:"jsonSchema"`
	SupportedSyncModes      []string               `json:"supportedSyncModes"`
	SourceDefinedCursor     bool                   `json:"sourceDefinedCursor"`
	DefaultCursorField      []string               `json:"defaultCursorField"`
	SourceDefinedPrimaryKey [][]string             `json:"sourceDefinedPrimaryKey"`
	Namespace               string                 `json:"namespace"`
}

type StreamConfigDetail struct {
	SyncMode            string     `json:"syncMode"`
	CursorField         []string   `json:"cursorField"`
	DestinationSyncMode string     `json:"destinationSyncMode"`
	PrimaryKey          [][]string `json:"primaryKey"`
	Selected            bool       `json:"selected"`
}

type ScheduleData struct {
	BasicSchedule BasicSchedule `json:"basicSchedule"`
}

type BasicSchedule struct {
	TimeUnit string `json:"timeUnit"`
	Units    int    `json:"units"`
}

// NewAirbyteClient cria um novo cliente Airbyte com timeouts robustos
func NewAirbyteClient(baseURL string) *AirbyteClient {
	return &AirbyteClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// WaitForReady verifica se o Airbyte está pronto com retries
func (c *AirbyteClient) WaitForReady() error {
	const maxRetries = 30
	const retryInterval = 5 * time.Second

	healthURL := c.BaseURL + "/api/v1/health"

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", healthURL, nil)
		if err != nil {
			return fmt.Errorf("creating request failed: %v", err)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			fmt.Printf("Health check attempt %d failed: %v\n", i+1, err)
			time.Sleep(retryInterval)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("✅ Airbyte is ready and healthy")
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Health check attempt %d: status %d, body: %s\n", i+1, resp.StatusCode, string(body))

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("airbyte not ready after %d attempts", maxRetries)
}

// GetFirstWorkspace obtém o primeiro workspace disponível
func (c *AirbyteClient) GetFirstWorkspace() (string, error) {
	resp, err := c.makeRequest("POST", "/api/v1/workspaces/list", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list workspaces: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("workspace list API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Workspaces []struct {
			WorkspaceID string `json:"workspaceId"`
			Name        string `json:"name"`
		} `json:"workspaces"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding workspace response failed: %v", err)
	}

	if len(result.Workspaces) == 0 {
		return "", fmt.Errorf("no workspaces found")
	}

	fmt.Printf("✅ Found workspace: %s (%s)\n", result.Workspaces[0].Name, result.Workspaces[0].WorkspaceID)
	return result.Workspaces[0].WorkspaceID, nil
}

// CreateSource cria uma nova source no Airbyte
func (c *AirbyteClient) CreateSource(workspaceID, name, sourceDefinitionID string, config map[string]interface{}) (string, error) {
	sourceReq := map[string]interface{}{
		"workspaceId":             workspaceID,
		"name":                    name,
		"sourceDefinitionId":      sourceDefinitionID,
		"connectionConfiguration": config,
	}

	resp, err := c.makeRequest("POST", "/api/v1/sources/create", sourceReq)
	if err != nil {
		return "", fmt.Errorf("failed to create source: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("source creation API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SourceID string `json:"sourceId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding source response failed: %v", err)
	}

	fmt.Printf("✅ Created source: %s (%s)\n", name, result.SourceID)
	return result.SourceID, nil
}

// CreateDestination cria um novo destination no Airbyte
func (c *AirbyteClient) CreateDestination(workspaceID, name, destinationDefinitionID string, config map[string]interface{}) (string, error) {
	destinationReq := map[string]interface{}{
		"workspaceId":             workspaceID,
		"name":                    name,
		"destinationDefinitionId": destinationDefinitionID,
		"connectionConfiguration": config,
	}

	resp, err := c.makeRequest("POST", "/api/v1/destinations/create", destinationReq)
	if err != nil {
		return "", fmt.Errorf("failed to create destination: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("destination creation API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		DestinationID string `json:"destinationId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding destination response failed: %v", err)
	}

	fmt.Printf("✅ Created destination: %s (%s)\n", name, result.DestinationID)
	return result.DestinationID, nil
}

// CreateConnection cria uma conexão entre source e destination
func (c *AirbyteClient) CreateConnection(sourceID, destinationID, name string) (string, error) {
	connectionConfig := map[string]interface{}{
		"name":          name,
		"sourceId":      sourceID,
		"destinationId": destinationID,
		"syncCatalog": map[string]interface{}{
			"streams": []map[string]interface{}{
				{
					"stream": map[string]interface{}{
						"name": "breweries",
						"jsonSchema": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
						"supportedSyncModes":      []string{"full_refresh", "incremental"},
						"sourceDefinedCursor":     false,
						"defaultCursorField":      []string{},
						"sourceDefinedPrimaryKey": [][]string{{"id"}},
						"namespace":               "public",
					},
					"config": map[string]interface{}{
						"syncMode":            "full_refresh",
						"cursorField":         []string{},
						"destinationSyncMode": "append",
						"primaryKey":          [][]string{{"id"}},
						"selected":            true,
						"aliasName":           "breweries",
					},
				},
			},
		},
		"scheduleType": "manual",
		"scheduleData": map[string]interface{}{
			"basicSchedule": map[string]interface{}{
				"timeUnit": "hours",
				"units":    24,
			},
		},
		"status": "active",
	}

	resp, err := c.makeRequest("POST", "/api/v1/connections/create", connectionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create connection: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("connection creation API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ConnectionID string `json:"connectionId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding connection response failed: %v", err)
	}

	fmt.Printf("✅ Created connection: %s (%s)\n", name, result.ConnectionID)
	return result.ConnectionID, nil
}

// makeRequest helper method para fazer requisições HTTP
func (c *AirbyteClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	// Optionally log the request here if needed

	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body failed: %v", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	fmt.Printf("📡 Response status: %d\n", resp.StatusCode)
	return resp, nil
}

// TestConnection testa uma conexão existente
func (c *AirbyteClient) TestConnection(connectionID string) error {
	testReq := map[string]interface{}{
		"connectionId": connectionID,
	}

	resp, err := c.makeRequest("POST", "/api/v1/connections/get", testReq)
	if err != nil {
		return fmt.Errorf("failed to test connection: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed with status %d", resp.StatusCode)
	}

	fmt.Printf("✅ Connection %s is valid\n", connectionID)
	return nil
}

// SyncConnection inicia uma sincronização manual
func (c *AirbyteClient) SyncConnection(connectionID string) error {
	syncReq := map[string]interface{}{
		"connectionId": connectionID,
	}

	resp, err := c.makeRequest("POST", "/api/v1/connections/sync", syncReq)
	if err != nil {
		return fmt.Errorf("failed to start sync: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ Started sync for connection: %s\n", connectionID)
	return nil
}
