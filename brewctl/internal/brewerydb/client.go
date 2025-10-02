package brewerydb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BreweryDBClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Brewery struct {
	ID             string `json:"id" bson:"id"`
	Name           string `json:"name" bson:"name"`
	BreweryType    string `json:"brewery_type" bson:"brewery_type"`
	Street         string `json:"street" bson:"street"`
	Address1       string `json:"address_1" bson:"address_1"`
	Address2       string `json:"address_2" bson:"address_2"`
	Address3       string `json:"address_3" bson:"address_3"`
	City           string `json:"city" bson:"city"`
	State          string `json:"state" bson:"state"`
	CountyProvince string `json:"county_province" bson:"county_province"`
	PostalCode     string `json:"postal_code" bson:"postal_code"`
	Country        string `json:"country" bson:"country"`
	Longitude      string `json:"longitude" bson:"longitude"`
	Latitude       string `json:"latitude" bson:"latitude"`
	Phone          string `json:"phone" bson:"phone"`
	WebsiteURL     string `json:"website_url" bson:"website_url"`
	UpdatedAt      string `json:"updated_at" bson:"updated_at"`
	CreatedAt      string `json:"created_at" bson:"created_at"`
}

type Meta struct {
	Total   string `json:"total"`
	Page    string `json:"page"`
	PerPage string `json:"per_page"`
}

type BreweryResponse struct {
	Data []Brewery `json:"data"`
	Meta Meta      `json:"meta,omitempty"`
}

func NewBreweryDBClient() *BreweryDBClient {
	return &BreweryDBClient{
		BaseURL: "https://api.openbrewerydb.org/v1",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAllBreweries - Busca TODAS as cervejarias com paginação
func (c *BreweryDBClient) GetAllBreweries() ([]Brewery, error) {
	var allBreweries []Brewery
	page := 1
	perPage := 200 // Máximo permitido pela API

	for {
		fmt.Printf("📋 Buscando página %d de cervejarias...\n", page)

		url := fmt.Sprintf("%s/breweries?page=%d&per_page=%d", c.BaseURL, page, perPage)
		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status code inválido: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta: %v", err)
		}

		var breweryResp BreweryResponse
		if err := json.Unmarshal(body, &breweryResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
		}

		if len(breweryResp.Data) == 0 {
			break // Última página
		}

		allBreweries = append(allBreweries, breweryResp.Data...)
		fmt.Printf("✅ Página %d: %d cervejarias (Total: %d)\n", page, len(breweryResp.Data), len(allBreweries))

		// Verifica se chegou na última página
		if len(breweryResp.Data) < perPage {
			break
		}

		page++
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	return allBreweries, nil
}

// GetBreweriesByCity - Busca cervejarias por cidade
func (c *BreweryDBClient) GetBreweriesByCity(city string) ([]Brewery, error) {
	url := fmt.Sprintf("%s/breweries?by_city=%s&per_page=50", c.BaseURL, city)
	return c.makeRequest(url)
}

// GetBreweriesByState - Busca cervejarias por estado
func (c *BreweryDBClient) GetBreweriesByState(state string) ([]Brewery, error) {
	url := fmt.Sprintf("%s/breweries?by_state=%s&per_page=50", c.BaseURL, state)
	return c.makeRequest(url)
}

// GetBreweriesByType - Busca cervejarias por tipo
func (c *BreweryDBClient) GetBreweriesByType(breweryType string) ([]Brewery, error) {
	url := fmt.Sprintf("%s/breweries?by_type=%s&per_page=50", c.BaseURL, breweryType)
	return c.makeRequest(url)
}

// GetRandomBreweries - Busca cervejarias aleatórias
func (c *BreweryDBClient) GetRandomBreweries(size int) ([]Brewery, error) {
	url := fmt.Sprintf("%s/breweries/random?size=%d", c.BaseURL, size)
	return c.makeRequest(url)
}

// SearchBreweries - Busca cervejarias por termo
func (c *BreweryDBClient) SearchBreweries(query string) ([]Brewery, error) {
	url := fmt.Sprintf("%s/breweries/search?query=%s", c.BaseURL, query)
	return c.makeRequest(url)
}

// GetBreweryByID - Busca cervejaria específica por ID
func (c *BreweryDBClient) GetBreweryByID(id string) (*Brewery, error) {
	url := fmt.Sprintf("%s/breweries/%s", c.BaseURL, id)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var brewery Brewery
	if err := json.Unmarshal(body, &brewery); err != nil {
		return nil, err
	}

	return &brewery, nil
}

// GetMetadata - Busca metadados para validação
func (c *BreweryDBClient) GetMetadata() (*Meta, error) {
	url := fmt.Sprintf("%s/breweries/meta", c.BaseURL)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var meta Meta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (c *BreweryDBClient) makeRequest(url string) ([]Brewery, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var breweries []Brewery
	if err := json.Unmarshal(body, &breweries); err != nil {
		return nil, err
	}

	return breweries, nil
}
