package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// ProviderConfig specifies the OAuth 2.0 provider configuration endpoints and parameters.
type ProviderConfig struct {
	ClientID        string
	ClientSecret    string
	AuthURL         string
	TokenURL        string
	UserInfoURL     string
	Scopes          []string
	ExtraAuthParams map[string]string
}

// TokenResult represents the resolved credentials from OAuth 2.0 PKCE exchange.
type TokenResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	IDToken      string    `json:"id_token,omitempty"`
	UserEmail    string    `json:"user_email,omitempty"`
}

// ExchangeCode performs an authorization_code token exchange using the PKCE code verifier.
func ExchangeCode(ctx context.Context, cfg ProviderConfig, code, verifier, redirectURI string) (*TokenResult, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   cfg.AuthURL,
			TokenURL:  cfg.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectURI,
		Scopes:      cfg.Scopes,
	}

	tok, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}

	res := &TokenResult{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		ExpiresAt:    tok.Expiry,
	}
	if idTok, ok := tok.Extra("id_token").(string); ok {
		res.IDToken = idTok
	}

	if cfg.UserInfoURL != "" && res.AccessToken != "" {
		client := oauthConfig.Client(ctx, tok)
		client.Timeout = 30 * time.Second
		req, err := http.NewRequestWithContext(ctx, "GET", cfg.UserInfoURL, nil)
		if err == nil && req != nil {
			if uResp, err := client.Do(req); err == nil {
				defer uResp.Body.Close()
				if uResp.StatusCode == http.StatusOK {
					var userinfo struct {
						Email string `json:"email"`
					}
					if err := json.NewDecoder(io.LimitReader(uResp.Body, 1<<20)).Decode(&userinfo); err == nil {
						res.UserEmail = userinfo.Email
					}
				}
			}
		}
	}
	return res, nil
}
