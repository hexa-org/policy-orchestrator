package openpolicyagent

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"gopkg.in/square/go-jose.v2/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GithubHTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	Do(request *http.Request) (resp *http.Response, err error)
}

type GithubBundleClientOptions struct {
	HTTPClient GithubHTTPClient
}

type githubCredentialsKey struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

type GithubBundleClient struct {
	account        string
	repo           string
	bundlePath     string
	credentialsKey githubCredentialsKey
	httpClient     GithubHTTPClient
}

type GithubPublishInfo struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Sha     string `json:"sha,omitempty"`
}

type githubContentsResponse struct {
	Content string `json:"content"`
	Sha     string `json:"sha,omitempty"`
}

var errNotFound = errors.New("unable to read bundle from Github: NOT FOUND")

func (o *GithubBundleClientOptions) deleteMe_WithHttpClient(client GithubHTTPClient) {
	o.HTTPClient = client
}

func (o *GithubBundleClientOptions) githubHttpClient() GithubHTTPClient {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}

	return &http.Client{
		Timeout: time.Duration(3) * time.Second,
	}
}

func NewGithubBundleClient(account, repo, bundlePath string, key []byte, opts GithubBundleClientOptions) (*GithubBundleClient, error) {

	if len(account) == 0 || len(repo) == 0 || len(bundlePath) == 0 {
		return nil, fmt.Errorf("required config: account, repo, branch, bundle")
	}
	var ghCredentials githubCredentialsKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&ghCredentials)
	if err != nil {
		return nil, err
	}

	if err := validator.New().Struct(ghCredentials); err != nil {
		return nil, err
	}

	return &GithubBundleClient{
		account:        account,
		repo:           repo,
		bundlePath:     bundlePath,
		credentialsKey: ghCredentials,
		httpClient:     opts.githubHttpClient(),
	}, nil
}

func (g *GithubBundleClient) GetDataFromBundle(path string) ([]byte, error) {
	metadataUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		g.account, g.repo, g.bundlePath)

	contentResp, err := g.fetchBundle(metadataUrl)
	if err != nil {
		return nil, err
	}

	content, _ := base64.StdEncoding.DecodeString(contentResp.Content)
	gz, gzipErr := compressionsupport.UnGzip(bytes.NewReader(content))
	if gzipErr != nil {
		return nil, gzipErr
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, tarErr
	}

	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

func (g *GithubBundleClient) PostBundle(bundle []byte) (int, error) {
	metadataUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		g.account, g.repo, g.bundlePath)
	contentResp, err := g.fetchBundle(metadataUrl)
	if err != nil && !errors.Is(err, errNotFound) {
		return 0, err
	}

	var sha string
	if contentResp != nil {
		sha = contentResp.Sha
	}

	requestBody, err := newGithubRequestBody(bundle, sha)
	if err != nil {
		return 0, err
	}

	uploadResponse, err := g.newRequest(http.MethodPut, g.credentialsKey.AccessToken, metadataUrl, requestBody)
	defer uploadResponse.Body.Close()
	responseBody, err := io.ReadAll(uploadResponse.Body)
	if err != nil {
		return 0, err
	}

	if uploadResponse.StatusCode != http.StatusCreated && uploadResponse.StatusCode != http.StatusOK {
		errorMessage := parseGithubError(responseBody)
		return uploadResponse.StatusCode, fmt.Errorf("unable to write bundle to Github: %s", errorMessage)
	}

	return http.StatusCreated, nil
}

func (g *GithubBundleClient) fetchBundle(url string) (*githubContentsResponse, error) {
	resp, err := g.newRequest(http.MethodGet, g.credentialsKey.AccessToken, url, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}

	if resp.StatusCode != http.StatusOK {
		errorMessage := parseGithubError(body)
		return nil, fmt.Errorf("unable to read bundle from Github: %s", errorMessage)
	}

	contentResp := githubContentsResponse{}
	err = json.Unmarshal(body, &contentResp)
	if err != nil {
		return nil, err
	}

	return &contentResp, nil
}

func parseGithubError(body []byte) string {
	parsedErrorResp := struct {
		Message string `json:"message"`
	}{}
	err := json.Unmarshal(body, &parsedErrorResp)
	if err != nil {
		return err.Error()
	}
	return parsedErrorResp.Message
}

func (g *GithubBundleClient) newRequest(method, token, url string, body io.ReadCloser) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	//req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return g.httpClient.Do(req)
}

func newGithubRequestBody(content []byte, sha string) (io.ReadCloser, error) {
	publishInfo, err := json.Marshal(&GithubPublishInfo{
		Message: "Update opa bundle",
		Content: base64.StdEncoding.EncodeToString(content),
		Sha:     sha,
	})

	if err != nil {
		return nil, err
	}

	return io.NopCloser(strings.NewReader(string(publishInfo))), nil
}
