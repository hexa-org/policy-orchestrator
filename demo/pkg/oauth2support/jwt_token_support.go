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
	EnvOAuthJwksUrl       string = "HEXA_TOKEN_JWKSURL"
	EnvJwtAuth            string = "HEXA_JWT_AUTH_ENABLE"
	EnvJwtRealm           string = "HEXA_JWT_REALM"
	EnvOAuthClientId      string = "HEXA_OAUTH_CLIENT_ID"
	EnvOAuthClientSecret  string = "HEXA_OAUTH_CLIENT_SECRET"
	EnvOAuthClientScope   string = "HEXA_OAUTH_CLIENT_SCOPE"
	EnvOAuthTokenEndpoint string = "HEXA_OAUTH_TOKEN_ENDPOINT"
)

type ResourceServerJwtHandler struct {
	jwksUrl string
	realm   string
	enable  bool
	Key     keyfunc.Keyfunc
}

func NewResourceServerJwtHandler() *ResourceServerJwtHandler {
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
			return &ResourceServerJwtHandler{
				jwksUrl: url,
				enable:  true,
				Key:     jwkKeyfunc,
				realm:   realm,
			}
		}
		log.Fatalf("Configuration parameter %s not set", EnvOAuthJwksUrl)
	}
	log.Println("JWT Authentication disabled.")
	return &ResourceServerJwtHandler{enable: false}
}

func JwtAuthenticationHandler(next http.HandlerFunc, s *ResourceServerJwtHandler) http.HandlerFunc {
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

func (s *ResourceServerJwtHandler) authenticate(w http.ResponseWriter, r *http.Request) (*jwt.Token, bool) {
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
		token, err := jwt.Parse(tokenString, s.Key.Keyfunc)

		if err != nil {
			headerMsg := fmt.Sprintf("Bearer realm=\"%s\", error=\"invalid_token\", error_description=\"%s\"", s.realm, err.Error())
			w.Header().Set("www-authenticate", headerMsg)
			w.WriteHeader(http.StatusUnauthorized)
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
