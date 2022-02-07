package amazonsupport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwk"
	"net/http"
)

type AmazonCognitoConfiguration struct {
	Region       string
	Domain       string
	RedirectUrl  string
	UserPoolId       string
	UserPoolClientId     string
	UserPoolClientSecret string
}

type AmazonCognitoToken struct {
	Token   string `json:"id_token"`
	Type    string `json:"token_type"`
	Expires int32  `json:"expires_in"`
}

type AmazonCognitoClaims struct {
	Email string `json:email`
	jwt.StandardClaims
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AmazonSupport struct {
	client       HTTPClient
	amazonConfig AmazonCognitoConfiguration
	session      *sessions.CookieStore
}

func NewAmazonSupport(client HTTPClient, amazonConfig AmazonCognitoConfiguration, session *sessions.CookieStore) *AmazonSupport {
	return &AmazonSupport{client, amazonConfig, session}
}

func (a *AmazonSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code := r.URL.Query().Get("code"); code != "" {

			clientInfo := []byte(fmt.Sprintf("%s:%s", a.amazonConfig.UserPoolClientId, a.amazonConfig.UserPoolClientSecret))
			clientInfoBase64 := base64.StdEncoding.EncodeToString(clientInfo)
			domain := fmt.Sprintf("%s.auth.%s.amazoncognito.com", a.amazonConfig.Domain, a.amazonConfig.Region)
			url := fmt.Sprintf("https://%s/oauth2/token?code=%s&grant_type=authorization_code&redirect_uri=%s",
				domain, code, a.amazonConfig.RedirectUrl)
			post, postError := a.requestToken(url, clientInfoBase64)
			if postError != nil {
				return
			}

			var token AmazonCognitoToken
			if json.NewDecoder(post.Body).Decode(&token) != nil {
				return
			}

			/// todo -

			publicKeysURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", a.amazonConfig.Region, a.amazonConfig.UserPoolId)
			publicKeySet, _ := jwk.Fetch(context.Background(), publicKeysURL)
			claims := &AmazonCognitoClaims{}
			_, _ = jwt.ParseWithClaims(token.Token, claims, func(token *jwt.Token) (interface{}, error) {
				claims = token.Claims.(*AmazonCognitoClaims)
				keys, _ := publicKeySet.LookupKeyID(token.Header["kid"].(string))
				var tokenKey interface{}
				keysErr := keys.Raw(&tokenKey)
				return tokenKey, keysErr
			})
			session, _ := a.session.Get(r, "session")
			session.Values["principal"] = claims.Email
			_ = session.Save(r, w)
		}
		next.ServeHTTP(w, r)
	})

}

func (a *AmazonSupport) requestToken(url string, clientInfoBase64 string) (*http.Response, error) {
	request, _ := http.NewRequest("POST", url, nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", clientInfoBase64))
	return a.client.Do(request)
}
