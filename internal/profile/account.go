package profile

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// AccountInfo contains identity and authentication metadata for a profile.
type AccountInfo struct {
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
	AuthMethod string `json:"auth_method,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
}

// GetProfileAccountInfo extracts the user email, display name, and auth method
// associated with a profile's saved credentials.
func GetProfileAccountInfo(profileDir string) AccountInfo {
	return GetProfileAccountInfoForAgent(profileDir, "")
}

// GetProfileAccountInfoForAgent extracts identity and auth metadata for a profile,
// prioritizing the specified agent if provided.
func GetProfileAccountInfoForAgent(profileDir, agentName string) AccountInfo {
	if profileDir == "" {
		return AccountInfo{}
	}

	switch agentName {
	case "codex", "codex-cli", "openai-codex":
		codexAuthPath := filepath.Join(profileDir, ".codex", "auth.json")
		if info := extractFromCodexAuthFile(codexAuthPath); info.Email != "" || info.AuthMethod != "" {
			return info
		}
	case "gemini", "gemini-cli":
		geminiTokenPath := filepath.Join(profileDir, ".gemini", "gemini-oauth-token")
		if info := extractFromTokenFile(geminiTokenPath); info.Email != "" || info.AuthMethod != "" {
			return info
		}
	case "agy", "antigravity":
		agyTokenPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
		if info := extractFromTokenFile(agyTokenPath); info.Email != "" || info.AuthMethod != "" {
			return info
		}
	}

	// 1. Check Antigravity OAuth token
	agyTokenPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if info := extractFromTokenFile(agyTokenPath); info.Email != "" || info.AuthMethod != "" {
		return info
	}

	// 2. Check Gemini CLI OAuth token
	geminiTokenPath := filepath.Join(profileDir, ".gemini", "gemini-oauth-token")
	if info := extractFromTokenFile(geminiTokenPath); info.Email != "" || info.AuthMethod != "" {
		return info
	}

	// 3. Check Google Application Default Credentials (ADC)
	adcPath := filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json")
	if info := extractFromADCFile(adcPath); info.Email != "" || info.AuthMethod != "" {
		return info
	}

	// 4. Check Codex auth file
	codexAuthPath := filepath.Join(profileDir, ".codex", "auth.json")
	if info := extractFromCodexAuthFile(codexAuthPath); info.Email != "" || info.AuthMethod != "" {
		return info
	}

	return AccountInfo{}
}

func extractFromTokenFile(path string) AccountInfo {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return AccountInfo{}
	}
	return ParseTokenAccountInfo(data)
}

// ParseTokenAccountInfo parses token JSON data and extracts email, name, auth method, and project ID.
func ParseTokenAccountInfo(data []byte) AccountInfo {
	var root struct {
		IDToken        string `json:"id_token"`
		AuthMethod     string `json:"auth_method"`
		UserEmail      string `json:"user_email"`
		Email          string `json:"email"`
		ClientEmail    string `json:"client_email"`
		ProjectID      string `json:"project_id"`
		QuotaProjectID string `json:"quota_project_id"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return AccountInfo{}
	}

	var email, name string
	if root.IDToken != "" {
		parts := strings.Split(root.IDToken, ".")
		if len(parts) >= 2 {
			payload := parts[1]
			decoded, err := base64.RawURLEncoding.DecodeString(payload)
			if err != nil {
				decoded, err = base64.URLEncoding.DecodeString(payload)
			}
			if err == nil {
				var claims struct {
					Email string `json:"email"`
					Name  string `json:"name"`
				}
				if err := json.Unmarshal(decoded, &claims); err == nil {
					email = claims.Email
					name = claims.Name
				}
			}
		}
	}

	if email == "" {
		if root.UserEmail != "" {
			email = root.UserEmail
		} else if root.Email != "" {
			email = root.Email
		} else if root.ClientEmail != "" {
			email = root.ClientEmail
		}
	}

	authMethod := "Google OAuth"
	if root.AuthMethod != "" {
		if root.AuthMethod == "consumer" {
			authMethod = "Google OAuth (consumer)"
		} else {
			authMethod = root.AuthMethod
		}
	}

	projID := root.ProjectID
	if projID == "" {
		projID = root.QuotaProjectID
	}

	return AccountInfo{
		Email:      email,
		Name:       name,
		AuthMethod: authMethod,
		ProjectID:  projID,
	}
}

func extractFromADCFile(path string) AccountInfo {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return AccountInfo{}
	}
	var root struct {
		ClientEmail    string `json:"client_email"`
		Type           string `json:"type"`
		ProjectID      string `json:"project_id"`
		QuotaProjectID string `json:"quota_project_id"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return AccountInfo{}
	}

	authMethod := "Google ADC"
	if root.Type != "" {
		authMethod = "Google ADC (" + root.Type + ")"
	}
	projID := root.ProjectID
	if projID == "" {
		projID = root.QuotaProjectID
	}
	return AccountInfo{
		Email:      root.ClientEmail,
		AuthMethod: authMethod,
		ProjectID:  projID,
	}
}

func extractFromCodexAuthFile(path string) AccountInfo {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return AccountInfo{}
	}
	return ParseCodexAuthAccountInfo(data)
}

// ParseCodexAuthAccountInfo extracts account metadata from OpenAI Codex auth.json data.
func ParseCodexAuthAccountInfo(data []byte) AccountInfo {
	var root struct {
		Tokens struct {
			IDToken     string `json:"id_token"`
			AccessToken string `json:"access_token"`
			AccountID   string `json:"account_id"`
		} `json:"tokens"`
		Email        string `json:"email"`
		UserEmail    string `json:"user_email"`
		APIKey       string `json:"api_key"`
		OpenAIAPIKey string `json:"openai_api_key"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return AccountInfo{}
	}

	var email, name, planType string
	if root.Tokens.IDToken != "" {
		parts := strings.Split(root.Tokens.IDToken, ".")
		if len(parts) >= 2 {
			payload := parts[1]
			decoded, err := base64.RawURLEncoding.DecodeString(payload)
			if err != nil {
				decoded, err = base64.URLEncoding.DecodeString(payload)
			}
			if err == nil {
				var claims struct {
					Email      string `json:"email"`
					Name       string `json:"name"`
					OpenAIAuth struct {
						ChatGPTPlanType string `json:"chatgpt_plan_type"`
						UserID          string `json:"user_id"`
					} `json:"https://api.openai.com/auth"`
				}
				if err := json.Unmarshal(decoded, &claims); err == nil {
					email = claims.Email
					name = claims.Name
					planType = claims.OpenAIAuth.ChatGPTPlanType
				}
			}
		}
	}

	if email == "" {
		if root.Email != "" {
			email = root.Email
		} else if root.UserEmail != "" {
			email = root.UserEmail
		}
	}

	authMethod := "OpenAI OAuth"
	if planType != "" && planType != "unknown" {
		switch strings.ToLower(planType) {
		case "plus":
			authMethod = "ChatGPT Plus"
		case "pro":
			authMethod = "ChatGPT Pro"
		case "team":
			authMethod = "ChatGPT Team"
		case "enterprise":
			authMethod = "ChatGPT Enterprise"
		case "free":
			authMethod = "ChatGPT Free"
		default:
			authMethod = "ChatGPT (" + planType + ")"
		}
	} else if root.APIKey != "" || root.OpenAIAPIKey != "" {
		authMethod = "OpenAI API Key"
	}

	return AccountInfo{
		Email:      email,
		Name:       name,
		AuthMethod: authMethod,
	}
}

