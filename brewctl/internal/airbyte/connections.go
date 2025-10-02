package airbyte

import (
	"fmt"
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

	// 6. Testar a conexão (opcional)
	if err := c.TestConnection(connectionID); err != nil {
		fmt.Printf("⚠️ Connection test warning: %v\n", err)
	} else {
		fmt.Printf("✅ Connection test passed: %s\n", connectionID)
	}

	fmt.Println("🎯 Airbyte connections setup completed successfully!")
	return nil
}

// CreateBrewerySource cria uma source para a BreweryDB API
func (c *AirbyteClient) CreateBrewerySource(workspaceID string) (string, error) {
	sourceConfig := map[string]interface{}{
		"url":         "https://api.openbrewerydb.org/v1/breweries",
		"http_method": "GET",
		"pagination": map[string]interface{}{
			"strategy":        "PageIncrement",
			"page_size":       50,
			"page_size_field": "per_page",
			"page_field":      "page",
		},
		"json_path": "$[*]",
		"headers": map[string]interface{}{
			"Accept": "application/json",
		},
	}

	// Source Definition ID para HTTP Request (valor padrão do Airbyte)
	sourceDefinitionID := "decd338e-5647-4c0b-adf4-da0e75f5a750"

	return c.CreateSource(workspaceID, "BreweryDB API", sourceDefinitionID, sourceConfig)
}

// CreateMongoDBDestination cria um destination para MongoDB
func (c *AirbyteClient) CreateMongoDBDestination(workspaceID string) (string, error) {
	destinationConfig := map[string]interface{}{
		"host":     "mongodb.default.svc.cluster.local",
		"port":     27017,
		"database": "breweries_db",
		"auth_type": map[string]interface{}{
			"authorization": "none",
		},
		"tunnel_method": map[string]interface{}{
			"tunnel_method": "NO_TUNNEL",
		},
	}

	// Destination Definition ID para MongoDB (valor padrão do Airbyte)
	destinationDefinitionID := "8e1c2c78-6c49-4c4a-b2c5-6e0b4b3c5a7b"

	return c.CreateDestination(workspaceID, "Breweries MongoDB", destinationDefinitionID, destinationConfig)
}
