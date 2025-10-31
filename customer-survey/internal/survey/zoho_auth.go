package survey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// ZohoConfig holds Zoho Creator API configuration
type ZohoConfig struct {
	AccountOwner string `json:"account_owner"`
	AppLinkName  string `json:"app_link_name"`
	FormLinkName string `json:"form_link_name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	DataCenter   string `json:"data_center"` // Default: "com" (US), "eu", "in", "com.au", "jp"
}

// ZohoAuth manages OAuth tokens for Zoho Creator API
type ZohoAuth struct {
	config       *ZohoConfig
	accessToken  string
	tokenExpiry  time.Time
	mu           sync.RWMutex
	tokenBaseURL string
}

// TokenResponse represents Zoho OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	APIServerID string `json:"api_domain"`
	TokenType   string `json:"token_type"`
}

// NewZohoAuth creates a new Zoho authentication manager
func NewZohoAuth(config *ZohoConfig) *ZohoAuth {
	if config.DataCenter == "" {
		config.DataCenter = "com" // Default to US
	}

	return &ZohoAuth{
		config:       config,
		tokenBaseURL: fmt.Sprintf("https://accounts.zoho.%s/oauth/v2/token", config.DataCenter),
	}
}

// LoadZohoConfig loads Zoho configuration from environment variables or config file
func LoadZohoConfig() (*ZohoConfig, error) {
	config := &ZohoConfig{}

	// Try environment variables first (most secure)
	config.AccountOwner = os.Getenv("ZOHO_ACCOUNT_OWNER")
	config.AppLinkName = os.Getenv("ZOHO_APP_LINK")
	config.FormLinkName = os.Getenv("ZOHO_FORM_LINK")
	config.ClientID = os.Getenv("ZOHO_CLIENT_ID")
	config.ClientSecret = os.Getenv("ZOHO_CLIENT_SECRET")
	config.RefreshToken = os.Getenv("ZOHO_REFRESH_TOKEN")
	config.DataCenter = os.Getenv("ZOHO_DATA_CENTER")

	// If environment variables are set, use them
	if config.ClientID != "" && config.ClientSecret != "" && config.RefreshToken != "" {
		return config, nil
	}

	// Otherwise, try loading from config file (less secure, but fallback)
	configPath := "configs/zoho_secure.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Zoho configuration not found. Set environment variables or create %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Zoho config: %w", err)
	}

	var fileConfig struct {
		Zoho ZohoConfig `json:"zoho"`
	}
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Zoho config: %w", err)
	}

	return &fileConfig.Zoho, nil
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (z *ZohoAuth) GetAccessToken() (string, error) {
	z.mu.RLock()
	// If token is valid and not expiring in next 5 minutes, return it
	if z.accessToken != "" && time.Now().Add(5*time.Minute).Before(z.tokenExpiry) {
		token := z.accessToken
		z.mu.RUnlock()
		return token, nil
	}
	z.mu.RUnlock()

	// Need to refresh token
	return z.refreshAccessToken()
}

// refreshAccessToken gets a new access token using the refresh token
func (z *ZohoAuth) refreshAccessToken() (string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Double-check after acquiring lock
	if z.accessToken != "" && time.Now().Add(5*time.Minute).Before(z.tokenExpiry) {
		return z.accessToken, nil
	}

	// Prepare refresh token request
	data := url.Values{}
	data.Set("refresh_token", z.config.RefreshToken)
	data.Set("client_id", z.config.ClientID)
	data.Set("client_secret", z.config.ClientSecret)
	data.Set("grant_type", "refresh_token")

	// Make request
	resp, err := http.PostForm(z.tokenBaseURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update stored token
	z.accessToken = tokenResp.AccessToken
	z.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return z.accessToken, nil
}

// GetAPIEndpoint returns the Zoho Creator API endpoint for form submission
func (z *ZohoAuth) GetAPIEndpoint() string {
	return fmt.Sprintf("https://creator.zoho.%s/api/v2/%s/%s/form/%s",
		z.config.DataCenter,
		z.config.AccountOwner,
		z.config.AppLinkName,
		z.config.FormLinkName,
	)
}

// SubmitToZohoCreator submits survey data to Zoho Creator using OAuth
func (z *ZohoAuth) SubmitToZohoCreator(data map[string]interface{}) error {
	// Get valid access token
	accessToken, err := z.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Prepare payload
	payload := map[string]interface{}{
		"data": data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create request
	apiURL := z.GetAPIEndpoint()
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit to Zoho: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Zoho API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Log success
	logPath := getLogPath("zoho_api.log")
	appendFile(logPath, fmt.Sprintf("%s | SUCCESS | Status: %d | Response: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), resp.StatusCode, string(body)))

	return nil
}
