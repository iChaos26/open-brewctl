package airbyte

import (
	"fmt"
	"io"
	"net/http"
)

// SetupConnections configura todas as conexões do Airbyte
func (c *AirbyteClient) SetupConnections() error {
	fmt.Println("🔗 Setting up Airbyte connections...")

	// 1. Aguardar Airbyte ficar pronto
	if err := c.WaitForReady(); err != nil {
		return fmt.Errorf("airbyte not ready: %v", err)
	}

	// 2. Obter workspace ID primeiro
	workspaceID, err := c.GetFirstWorkspace()
	if err != nil {
		return fmt.Errorf("failed to get workspace: %v", err)
	}
	fmt.Printf("✅ Using workspace ID: %s\n", workspaceID)

	// 3. Criar source da BreweryDB
	sourceID, err := c.CreateBrewerySource(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to create source: %v", err)
	}

	// 4. Criar destination do MongoDB
	destinationID, err := c.CreateMongoDBDestination(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to create destination: %v", err)
	}

	// 5. Criar conexão entre source e destination
	connectionID, err := c.CreateConnection(sourceID, destinationID, "BreweryDB to MongoDB Pipeline")
	if err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}

	// 6. Testar e iniciar a sincronização
	if err := c.TestAndSyncConnection(connectionID); err != nil {
		return fmt.Errorf("failed to sync connection: %v", err)
	}

	fmt.Println("🎯 Airbyte connections setup completed successfully!")
	return nil
}

// CreateBrewerySource cria uma source para a BreweryDB API
func (c *AirbyteClient) CreateBrewerySource(workspaceID string) (string, error) {
	sourceConfig := map[string]interface{}{
		"url_base":    "https://api.openbrewerydb.org/v1/breweries",
		"http_method": "GET",
		"request_parameters": map[string]string{
			"per_page": "50",
		},
		"pagination_strategy": "PageIncrement",
		"page_size":           50,
		"page_size_field":     "per_page",
		"page_field":          "page",
		"start_page":          1,
	}

	// CORREÇÃO: Source Definition ID correto para HTTP Request
	sourceDefinitionID := "8be1cf83-fde1-477f-a4ad-318d23c9f3c6"

	return c.CreateSource(workspaceID, "BreweryDB API", sourceDefinitionID, sourceConfig)
}

// CreateMongoDBDestination cria um destination para MongoDB
func (c *AirbyteClient) CreateMongoDBDestination(workspaceID string) (string, error) {
	destinationConfig := map[string]interface{}{
		"instance_type": "standalone",
		"host":          "mongodb.default.svc.cluster.local",
		"port":          27017,
		"database":      "breweries_db",
		"auth_type": map[string]interface{}{
			"authorization": "none",
		},
		"tls": false,
	}

	// CORREÇÃO: Destination Definition ID correto para MongoDB
	destinationDefinitionID := "8e1c2c78-6c49-4c4a-b2c5-6e0b4b3c5a7b"

	return c.CreateDestination(workspaceID, "Breweries MongoDB", destinationDefinitionID, destinationConfig)
}

// TestAndSyncConnection testa e inicia a sincronização
func (c *AirbyteClient) TestAndSyncConnection(connectionID string) error {
	fmt.Printf("🔍 Testing connection %s...\n", connectionID)

	// Primeiro testar a conexão
	testReq := map[string]interface{}{
		"connectionId": connectionID,
	}

	resp, err := c.makeRequest("POST", "/api/v1/connections/get", testReq)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connection test failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ Connection %s is valid\n", connectionID)

	// Iniciar sincronização
	fmt.Printf("🔄 Starting sync for connection %s...\n", connectionID)
	if err := c.SyncConnection(connectionID); err != nil {
		return fmt.Errorf("failed to start sync: %v", err)
	}

	return nil
}
