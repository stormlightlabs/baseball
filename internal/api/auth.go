package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"stormlightlabs.org/baseball/internal/core"
)

// AuthRoutes handles authentication endpoints
type AuthRoutes struct {
	userRepo   core.UserRepository
	tokenRepo  core.OAuthTokenRepository
	apiKeyRepo core.APIKeyRepository

	githubConfig   *oauth2.Config
	codebergConfig *oauth2.Config
}

func newGithubConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/github/callback"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func newCodebergConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     getEnv("CODEBERG_CLIENT_ID", ""),
		ClientSecret: getEnv("CODEBERG_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("CODEBERG_REDIRECT_URL", "http://localhost:8080/v1/auth/codeberg/callback"),
		Scopes:       []string{"read:user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://codeberg.org/login/oauth/authorize",
			TokenURL: "https://codeberg.org/login/oauth/access_token",
		},
	}
}

// NewAuthRoutes creates a new AuthRoutes instance
func NewAuthRoutes(userRepo core.UserRepository, tokenRepo core.OAuthTokenRepository, apiKeyRepo core.APIKeyRepository) *AuthRoutes {
	return &AuthRoutes{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		apiKeyRepo:     apiKeyRepo,
		githubConfig:   newGithubConf(),
		codebergConfig: newCodebergConf(),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RegisterRoutes registers all auth routes
func (r *AuthRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/auth/github", r.handleGitHubLogin)
	mux.HandleFunc("GET /v1/auth/github/callback", r.handleGitHubCallback)
	mux.HandleFunc("GET /v1/auth/codeberg", r.handleCodebergLogin)
	mux.HandleFunc("GET /v1/auth/codeberg/callback", r.handleCodebergCallback)
	mux.HandleFunc("POST /v1/auth/logout", r.handleLogout)
	mux.HandleFunc("GET /v1/auth/me", r.handleMe)
	mux.HandleFunc("POST /v1/auth/keys", r.handleCreateAPIKey)
	mux.HandleFunc("GET /v1/auth/keys", r.handleListAPIKeys)
	mux.HandleFunc("DELETE /v1/auth/keys/{id}", r.handleRevokeAPIKey)
}

// generateState creates a random state string for OAuth2 CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// handleGitHubLogin initiates GitHub OAuth flow
func (r *AuthRoutes) handleGitHubLogin(w http.ResponseWriter, req *http.Request) {
	state, err := generateState()
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to generate state: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   req.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	url := r.githubConfig.AuthCodeURL(state)
	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

// handleGitHubCallback processes GitHub OAuth callback
func (r *AuthRoutes) handleGitHubCallback(w http.ResponseWriter, req *http.Request) {
	state := req.URL.Query().Get("state")
	code := req.URL.Query().Get("code")

	cookie, err := req.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		writeInternalServerError(w, fmt.Errorf("invalid OAuth state"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	token, err := r.githubConfig.Exchange(req.Context(), code)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to exchange code: %w", err))
		return
	}

	user, err := r.getGitHubUser(req.Context(), token.AccessToken)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to get user: %w", err))
		return
	}

	dbUser, err := r.userRepo.GetByEmail(req.Context(), user.Email)
	if err != nil {
		dbUser, err = r.userRepo.Create(req.Context(), user.Email, &user.Name, &user.AvatarURL)
		if err != nil {
			writeInternalServerError(w, fmt.Errorf("failed to create user: %w", err))
			return
		}
	}

	if err := r.userRepo.UpdateLastLogin(req.Context(), dbUser.ID); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to update last login: %w", err))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	sessionToken, err := r.tokenRepo.Create(req.Context(), dbUser.ID, token.AccessToken, &token.RefreshToken, expiresAt)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to create session: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   req.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
}

// handleCodebergLogin initiates Codeberg OAuth flow
func (r *AuthRoutes) handleCodebergLogin(w http.ResponseWriter, req *http.Request) {
	state, err := generateState()
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to generate state: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   req.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	url := r.codebergConfig.AuthCodeURL(state)
	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

// handleCodebergCallback processes Codeberg OAuth callback
func (r *AuthRoutes) handleCodebergCallback(w http.ResponseWriter, req *http.Request) {
	state := req.URL.Query().Get("state")
	code := req.URL.Query().Get("code")

	cookie, err := req.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		writeInternalServerError(w, fmt.Errorf("invalid OAuth state"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	token, err := r.codebergConfig.Exchange(req.Context(), code)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to exchange code: %w", err))
		return
	}

	user, err := r.getCodebergUser(req.Context(), token.AccessToken)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to get user: %w", err))
		return
	}

	dbUser, err := r.userRepo.GetByEmail(req.Context(), user.Email)
	if err != nil {
		dbUser, err = r.userRepo.Create(req.Context(), user.Email, &user.Name, &user.AvatarURL)
		if err != nil {
			writeInternalServerError(w, fmt.Errorf("failed to create user: %w", err))
			return
		}
	}

	if err := r.userRepo.UpdateLastLogin(req.Context(), dbUser.ID); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to update last login: %w", err))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	sessionToken, err := r.tokenRepo.Create(req.Context(), dbUser.ID, token.AccessToken, &token.RefreshToken, expiresAt)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to create session: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   req.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
}

// handleLogout logs out the current user
func (r *AuthRoutes) handleLogout(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session_token")
	if err == nil {
		token, err := r.tokenRepo.GetByAccessToken(req.Context(), cookie.Value)
		if err == nil {
			r.tokenRepo.Delete(req.Context(), token.ID)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// handleMe returns the current authenticated user
func (r *AuthRoutes) handleMe(w http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("user").(*core.User)
	if !ok {
		writeInternalServerError(w, fmt.Errorf("unauthorized"))
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleCreateAPIKey creates a new API key for the authenticated user
func (r *AuthRoutes) handleCreateAPIKey(w http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("user").(*core.User)
	if !ok {
		writeInternalServerError(w, fmt.Errorf("unauthorized"))
		return
	}

	var input struct {
		Name      *string    `json:"name"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeInternalServerError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	apiKey, key, err := r.apiKeyRepo.Create(req.Context(), user.ID, input.Name, input.ExpiresAt)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to create API key: %w", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"api_key": apiKey,
		"key":     key,
		"warning": "This key will only be shown once. Please save it securely.",
	})
}

// handleListAPIKeys lists all API keys for the authenticated user
func (r *AuthRoutes) handleListAPIKeys(w http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("user").(*core.User)
	if !ok {
		writeInternalServerError(w, fmt.Errorf("unauthorized"))
		return
	}

	keys, err := r.apiKeyRepo.ListByUser(req.Context(), user.ID)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to list API keys: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, keys)
}

// handleRevokeAPIKey revokes an API key
func (r *AuthRoutes) handleRevokeAPIKey(w http.ResponseWriter, req *http.Request) {
	user, ok := req.Context().Value("user").(*core.User)
	if !ok {
		writeInternalServerError(w, fmt.Errorf("unauthorized"))
		return
	}

	id := req.PathValue("id")
	if id == "" {
		writeInternalServerError(w, fmt.Errorf("missing key ID"))
		return
	}

	apiKey, err := r.apiKeyRepo.GetByID(req.Context(), id)
	if err != nil {
		writeInternalServerError(w, fmt.Errorf("API key not found: %w", err))
		return
	}

	if apiKey.UserID != user.ID {
		writeInternalServerError(w, fmt.Errorf("unauthorized"))
		return
	}

	if err := r.apiKeyRepo.Revoke(req.Context(), id); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to revoke API key: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "API key revoked"})
}

// OAuthUser represents user info from OAuth provider
type OAuthUser struct {
	Email     string
	Name      string
	AvatarURL string
}

// getGitHubUser fetches user info from GitHub API
func (r *AuthRoutes) getGitHubUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var ghUser struct {
		Login     string  `json:"login"`
		Name      *string `json:"name"`
		Email     *string `json:"email"`
		AvatarURL string  `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, err
	}

	email := ""
	if ghUser.Email != nil && *ghUser.Email != "" {
		email = *ghUser.Email
	} else {
		emails, err := r.getGitHubEmails(ctx, accessToken)
		if err == nil && len(emails) > 0 {
			for _, e := range emails {
				if e.Primary && e.Verified {
					email = e.Email
					break
				}
			}
			if email == "" && len(emails) > 0 {
				email = emails[0].Email
			}
		}
	}

	if email == "" {
		return nil, fmt.Errorf("no email found for GitHub user")
	}

	name := ghUser.Login
	if ghUser.Name != nil && *ghUser.Name != "" {
		name = *ghUser.Name
	}

	return &OAuthUser{
		Email:     email,
		Name:      name,
		AvatarURL: ghUser.AvatarURL,
	}, nil
}

// getGitHubEmails fetches user emails from GitHub API
func (r *AuthRoutes) getGitHubEmails(ctx context.Context, accessToken string) ([]struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}

	return emails, nil
}

// getCodebergUser fetches user info from Codeberg API
func (r *AuthRoutes) getCodebergUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://codeberg.org/api/v1/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Codeberg API error: %s", string(body))
	}

	var cbUser struct {
		Login     string  `json:"login"`
		FullName  *string `json:"full_name"`
		Email     string  `json:"email"`
		AvatarURL string  `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cbUser); err != nil {
		return nil, err
	}

	if cbUser.Email == "" {
		return nil, fmt.Errorf("no email found for Codeberg user")
	}

	name := cbUser.Login
	if cbUser.FullName != nil && *cbUser.FullName != "" {
		name = *cbUser.FullName
	}

	return &OAuthUser{
		Email:     cbUser.Email,
		Name:      name,
		AvatarURL: cbUser.AvatarURL,
	}, nil
}

// AuthContext represents authentication context
type AuthContext struct {
	User   *core.User
	APIKey *core.APIKey
}

// AuthMiddleware provides authentication middleware that checks for session tokens or API keys
func AuthMiddleware(userRepo core.UserRepository, tokenRepo core.OAuthTokenRepository, apiKeyRepo core.APIKeyRepository, debugMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if debugMode {
				next.ServeHTTP(w, r)
				return
			}

			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 {
					credentials := parts[1]

					if scheme := strings.ToLower(parts[0]); scheme == "bearer" {
						if token, err := tokenRepo.GetByAccessToken(r.Context(), credentials); err == nil {
							if user, err := userRepo.GetByID(r.Context(), token.UserID); err == nil && user.IsActive {
								ctx := context.WithValue(r.Context(), "user", user)
								next.ServeHTTP(w, r.WithContext(ctx))
								return
							}
						}

						// TODO: pass API key as X-API-KEY header
						if apiKey, err := apiKeyRepo.GetByKey(r.Context(), credentials); err == nil && apiKey.IsActive {
							if user, err := userRepo.GetByID(r.Context(), apiKey.UserID); err == nil && user.IsActive {
								apiKeyRepo.UpdateLastUsed(r.Context(), apiKey.ID)
								ctx := context.WithValue(r.Context(), "user", user)
								ctx = context.WithValue(ctx, "api_key", apiKey)
								next.ServeHTTP(w, r.WithContext(ctx))
								return
							}
						}
					}
				}
			}

			if cookie, err := r.Cookie("session_token"); err == nil {
				if token, err := tokenRepo.GetByAccessToken(r.Context(), cookie.Value); err == nil {
					if user, err := userRepo.GetByID(r.Context(), token.UserID); err == nil && user.IsActive {
						ctx := context.WithValue(r.Context(), "user", user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			writeInternalServerError(w, fmt.Errorf("unauthorized"))
		})
	}
}

// OptionalAuthMiddleware provides optional authentication that doesn't fail if no auth is present
func OptionalAuthMiddleware(userRepo core.UserRepository, tokenRepo core.OAuthTokenRepository, apiKeyRepo core.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					credentials := parts[1]

					if token, err := tokenRepo.GetByAccessToken(r.Context(), credentials); err == nil {
						if user, err := userRepo.GetByID(r.Context(), token.UserID); err == nil && user.IsActive {
							ctx := context.WithValue(r.Context(), "user", user)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}

					if apiKey, err := apiKeyRepo.GetByKey(r.Context(), credentials); err == nil && apiKey.IsActive {
						if user, err := userRepo.GetByID(r.Context(), apiKey.UserID); err == nil && user.IsActive {
							apiKeyRepo.UpdateLastUsed(r.Context(), apiKey.ID)
							ctx := context.WithValue(r.Context(), "user", user)
							ctx = context.WithValue(ctx, "api_key", apiKey)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}

			if cookie, err := r.Cookie("session_token"); err == nil {
				if token, err := tokenRepo.GetByAccessToken(r.Context(), cookie.Value); err == nil {
					if user, err := userRepo.GetByID(r.Context(), token.UserID); err == nil && user.IsActive {
						ctx := context.WithValue(r.Context(), "user", user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
