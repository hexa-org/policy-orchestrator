package oidctestsupport

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type FakeAuthServer struct {
	Issuer     string
	clientID   string
	claims     map[string]interface{}
	signingKey SigningKey
	Server     *httptest.Server
}

type SigningKey struct {
	key *rsa.PrivateKey
	id  string
}

func NewFakeAuthServer(clientID string, claims map[string]interface{}) *FakeAuthServer {
	k, _ := rsa.GenerateKey(rand.Reader, 2048)

	authServer := &FakeAuthServer{
		signingKey: SigningKey{
			key: k,
			id:  "some-key-id",
		},
		clientID: clientID,
		claims:   claims,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", authServer.handleWellKnown)
	mux.HandleFunc("/authorize", authServer.handleAuthorize)
	mux.HandleFunc("/token", authServer.handleToken)
	mux.HandleFunc("/jwks", authServer.handleJWKS)

	authServer.Server = httptest.NewServer(mux)
	authServer.Issuer = authServer.Server.URL
	return authServer
}

func (as *FakeAuthServer) handleWellKnown(rw http.ResponseWriter, r *http.Request) {
	wellKnownConfig := struct {
		Issuer                string `json:"issuer,omitempty"`
		AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
		TokenEndpoint         string `json:"token_endpoint,omitempty"`
		JWKSEndpoint          string `json:"jwks_uri,omitempty"`
	}{
		Issuer:                as.Issuer,
		AuthorizationEndpoint: fmt.Sprintf("%s/authorize", as.Issuer),
		TokenEndpoint:         fmt.Sprintf("%s/token", as.Issuer),
		JWKSEndpoint:          fmt.Sprintf("%s/jwks", as.Issuer),
	}
	_ = json.NewEncoder(rw).Encode(wellKnownConfig)
}

func (as *FakeAuthServer) handleAuthorize(rw http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	redirect := r.URL.Query().Get("redirect_uri")

	redirectURI, _ := url.Parse(redirect)
	q := redirectURI.Query()
	q.Set("code", "42")
	q.Set("state", state)
	redirectURI.RawQuery = q.Encode()

	// todo - inject a way to assert authorize request made by client?

	http.Redirect(rw, r, redirectURI.String(), http.StatusFound)
}

func (as *FakeAuthServer) handleToken(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if r.FormValue("code") != "42" {
		http.Error(rw, fmt.Sprintf("invalid authorization code"), http.StatusBadRequest)
		return
	}

	token, err := as.buildJWT()
	if err != nil {
		http.Error(rw, fmt.Sprintf("unable to build token: %s", err), http.StatusInternalServerError)
		return
	}

	resp := struct {
		AccessToken string `json:"access_token,omitempty"`
		TokenType   string `json:"token_type"`
		IDToken     string `json:"id_token,omitempty"`
	}{
		AccessToken: token,
		TokenType:   "Bearer",
		IDToken:     token,
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(resp)
}

func (as *FakeAuthServer) handleJWKS(rw http.ResponseWriter, _ *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       as.signingKey.key.Public(),
		KeyID:     as.signingKey.id,
		Algorithm: "RS256",
		Use:       "sig",
	}
	var keys []jose.JSONWebKey
	keys = append(keys, jwk)

	jwks := jose.JSONWebKeySet{
		Keys: keys,
	}
	rw.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(rw).Encode(jwks)
	if err != nil {
		log.Println("unable to encode jwks: ", err.Error())
	}
}

func (as *FakeAuthServer) buildJWT() (string, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(time.Minute)

	regClaims := jwt.Claims{
		Issuer:   as.Issuer,
		Audience: jwt.Audience{as.clientID},
		Subject:  as.clientID,
		Expiry:   jwt.NewNumericDate(expiresAt),
		IssuedAt: jwt.NewNumericDate(issuedAt),
	}

	sk := jose.SigningKey{Algorithm: jose.RS256, Key: as.signingKey.key}
	so := (&jose.SignerOptions{}).
		WithType("at+jwt").
		WithHeader("kid", as.signingKey.id)
	signer, err := jose.NewSigner(sk, so)
	if err != nil {
		return "", fmt.Errorf("failed to initialize signer: %w", err)
	}

	raw, err := jwt.Signed(signer).Claims(regClaims).Claims(as.claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign claims: %w", err)
	}

	return raw, nil
}
