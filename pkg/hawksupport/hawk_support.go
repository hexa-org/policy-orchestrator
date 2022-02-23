package hawksupport

import (
	"github.com/hiyosi/hawk"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
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

func HawkGet(client HTTPClient, id string, key string, url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	authorize(req, id, key, url, "GET")
	return client.Do(req)
}

func HawkPost(client HTTPClient, id string, key string, url string, body io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	authorize(req, id, key, url, "POST")
	return client.Do(req)
}

func authorize(req *http.Request, id string, key string, url string, method string) {
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
	header, _ := c.Header(method, url)
	req.Header.Set("authorization", header)
}
