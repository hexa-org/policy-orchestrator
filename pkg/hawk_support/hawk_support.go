package hawk_support

import (
	"fmt"
	"github.com/hiyosi/hawk"
	"log"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type credentialStore struct {
	key string
}

func NewCredentialStore(key string) hawk.CredentialStore {
	return &credentialStore{key}
}

func (c *credentialStore) GetCredential(id string) (*hawk.Credential, error) {
	return &hawk.Credential{
		ID:  id,
		Key: c.key,
		Alg: hawk.SHA256,
	}, nil
}

func HawkMiddleware(next http.HandlerFunc, credentialStore hawk.CredentialStore, hostPort string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := hawk.NewServer(credentialStore)
		s.AuthOption = &hawk.AuthOption{
			CustomHostPort: hostPort,
		}
		cred, err := s.Authenticate(r)
		if err != nil {
			fmt.Printf("Hawk authentication error %v\n", err)
			w.Header().Set("www-authenticate", "hawk")
			w.WriteHeader(401)
			return
		}
		opt := &hawk.Option{
			TimeStamp: time.Now().Unix(),
		}
		h, _ := s.Header(r, cred, opt)
		w.Header().Set("server-authorization", h)
		next(w, r)
	}
}

func HawkGet(client HTTPClient, id string, key string, uri string) (*http.Response, error) {
	c := hawk.NewClient(
		&hawk.Credential{
			ID:  id,
			Key: key,
			Alg: hawk.SHA256,
		},
		&hawk.Option{
			TimeStamp: time.Now().Unix(),
			Nonce:     "nonce",
		},
	)
	header, _ := c.Header("GET", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Printf("Request failed '%v'", err)
	}
	req.Header.Set("authorization", header)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Found hawk response error %v\n", err)
	}
	return resp, err
}
