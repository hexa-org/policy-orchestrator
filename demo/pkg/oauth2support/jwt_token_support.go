package oauth2support

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	EnvOAuthJwksUrl string = "HEXA_TOKEN_JWKSURL"
	EnvJwtAuth      string = "HEXA_JWT_AUTH_ENABLE"
	EnvJwtRealm     string = "HEXA_JWT_REALM"
	EnvJwtAudience  string = "HEXA_JWT_AUDIENCE"
	EnvJwtScope     string = "HEXA_JWT_SCOPE"

	EnvOAuthClientId      string = "HEXA_OAUTH_CLIENT_ID"
	EnvOAuthClientSecret  string = "HEXA_OAUTH_CLIENT_SECRET"
	EnvOAuthClientScope   string = "HEXA_OAUTH_CLIENT_SCOPE"
	EnvOAuthTokenEndpoint string = "HEXA_OAUTH_TOKEN_ENDPOINT"
)

type ResourceJwtAuthorizer struct {
	jwksUrl string
	realm   string
	enable  bool
	Key     keyfunc.Keyfunc
	Aud     string
	Scope   string
}

func NewResourceJwtAuthorizer() *ResourceJwtAuthorizer {
	enable := os.Getenv(EnvJwtAuth)
	if enable == "true" {
		url := os.Getenv(EnvOAuthJwksUrl)
		if url != "" {
			jwkKeyfunc, err := keyfunc.NewDefaultCtx(context.Background(), []string{url})
			if err != nil {
				log.Fatalf("Failed to create client JWK set. Error: %s", err)
			}
			realm := os.Getenv(EnvJwtRealm)
			if realm == "" {
				log.Println(fmt.Sprintf("Warning: realm environment value not set (%s)", EnvJwtRealm))
				realm = "UNDEFINED"
			}
			aud := os.Getenv(EnvJwtAudience)
			if aud == "" {
				log.Println(fmt.Sprintf("Warning: audience environment value not set (%s)", EnvJwtAudience))
				log.Println("Defaulting to aud=orchestrator")
				aud = "orchestrator"
			}
			scope := os.Getenv(EnvJwtScope)
			if scope == "" {
				log.Println(fmt.Sprintf("Warning: scope environment value not set (%s)", EnvOAuthClientScope))
				log.Println("Defaulting to scope=orchestrator")
				scope = "orchestrator"
			}
			return &ResourceJwtAuthorizer{
				jwksUrl: url,
				enable:  true,
				Key:     jwkKeyfunc,
				realm:   realm,
				Aud:     aud,
				Scope:   scope,
			}
		}
		log.Fatalf("Configuration parameter %s not set", EnvOAuthJwksUrl)

	}
	log.Println("JWT Authentication disabled.")
	return &ResourceJwtAuthorizer{enable: false}
}

func JwtAuthenticationHandler(next http.HandlerFunc, s *ResourceJwtAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.enable {
			if r.Header.Get("Authorization") == "" {
				log.Println("Request missing authorization header")
				w.Header().Set("www-authenticate", fmt.Sprintf("Bearer realm=\"%s\"", s.realm))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			cred, valid := s.authenticate(w, r)
			if !valid {
				// Error has already been encoded in the response
				return
			}

			// Encode the subject into the header
			subj, err := cred.Claims.GetSubject()
			if err == nil {
				r.Header.Set("X-Subject", subj)
			}
		}
		next(w, r)
	}
}

type AccessTokenInfo struct {
	*jwt.RegisteredClaims
	Scope string `json:"scope"`
}

func (s *ResourceJwtAuthorizer) authenticate(w http.ResponseWriter, r *http.Request) (*jwt.Token, bool) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		w.Header().Set("www-authenticate", fmt.Sprintf("Bearer realm=\"%s\"", s.realm))
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	parts := strings.Split(authorization, " ")
	if len(parts) < 2 {
		headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"invalid_token\", error_description=\"%s\"", s.realm, "Missing authorization type or value")
		w.Header().Set("www-authenticate", headerMsg)
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	if strings.EqualFold(parts[0], "bearer") {
		tokenString := strings.TrimSpace(parts[1])

		token, err := jwt.ParseWithClaims(tokenString, &AccessTokenInfo{}, s.Key.Keyfunc)

		if err != nil {
			headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"invalid_token\", error_description=\"%s\"", s.realm, err.Error())
			w.Header().Set("www-authenticate", headerMsg)
			w.WriteHeader(http.StatusUnauthorized)
			// log.Printf("Authorization invalid: [%s]\n", err.Error())
			return nil, false
		}

		// Check Audience
		audMatch := false
		var audStrings []string
		audStrings, err = token.Claims.GetAudience()
		if err != nil {
			log.Printf("Error parsing audience from token claims: %s", err.Error())
		}
		for _, aud := range audStrings {
			if strings.EqualFold(aud, s.Aud) {
				audMatch = true
			}
		}
		if !audMatch {
			headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"invalid_token\", error_description=\"invalid audience\"", s.realm)
			w.Header().Set("www-authenticate", headerMsg)
			w.WriteHeader(http.StatusUnauthorized)
			// log.Printf("Authorization invalid: [%s]\n", err.Error())
			return nil, false
		}

		scopeMatch := false
		var scopes []string
		atToken := token.Claims.(*AccessTokenInfo)
		scopeString := atToken.Scope
		scopes = strings.Split(scopeString, " ")
		if s.Scope != "" {
			for _, scope := range scopes {
				if strings.EqualFold(s.Scope, scope) {
					scopeMatch = true
				}
			}
		}

		if !scopeMatch {
			headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"insufficient_scope\", error_description=\"requires scope=%s\"", s.realm, s.Scope)
			w.Header().Set("www-authenticate", headerMsg)
			w.WriteHeader(http.StatusForbidden)
			// log.Printf("Authorization invalid: [%s]\n", err.Error())
			return nil, false
		}
		return token, true
	}

	headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"invalid_token\", error_description=\"%s\"", s.realm, "Bearer token required")
	w.Header().Set("www-authenticate", headerMsg)
	w.WriteHeader(http.StatusUnauthorized)
	return nil, false
}

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type jwtClient struct {
	State  string `json:"state"`
	Config *clientcredentials.Config
}

type JwtClientHandler interface {
	GetHttpClient() *http.Client
	GetToken() (*oauth2.Token, error)
}

/*
NewJwtClientHandler opens a new JwtClientHandler which allows an OAuth Client to make calls to a JWT protected
endpoint. Configuration parameters are pulled from environment variables.
*/
func NewJwtClientHandler() JwtClientHandler {
	clientId := os.Getenv(EnvOAuthClientId)
	secret := os.Getenv(EnvOAuthClientSecret)
	tokenUrl := os.Getenv(EnvOAuthTokenEndpoint)
	if tokenUrl == "" {
		log.Println(fmt.Sprintf("Error: Token endpoint (%s) not declared", EnvOAuthTokenEndpoint))
	}

	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		TokenURL:     tokenUrl,
		AuthStyle:    oauth2.AuthStyle(oauth2.AuthStyleAutoDetect),
	}

	return NewJwtClientHandlerWithConfig(config)
}

/*
NewJwtClientHandlerWithConfig opens a new JwtClientHandler which allows an OAuth Client to make calls to a JWT protected
endpoint. The `config` parameter specifies a client credential for the OAuth2 Client Credential Flow
*/
func NewJwtClientHandlerWithConfig(config *clientcredentials.Config) JwtClientHandler {
	return &jwtClient{
		Config: config,
	}
}

// GetHttpClient returns an http.Client object that can be used to make calls to protected services. The client
// automatically appends the authorization header and handles refresh with the OAuth Token Server as needed.
func (j *jwtClient) GetHttpClient() *http.Client {
	return j.Config.Client(context.Background())
}

// GetToken returns a token object providing access to access token and refresh token as needed.
func (j *jwtClient) GetToken() (*oauth2.Token, error) {
	return j.Config.TokenSource(context.Background()).Token()
}
