package amazonsupport

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strings"
)

type AmazonCognitoConfiguration struct {
	Region               string
	Domain               string
	RedirectUrl          string
	UserPoolId           string
	UserPoolClientId     string
	UserPoolClientSecret string
}

type AmazonCognitoClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AmazonSupport struct {
	client       HTTPClient
	amazonConfig AmazonCognitoConfiguration
	claimsParser ClaimsParser
	session      *sessions.CookieStore
}

func NewAmazonSupport(client HTTPClient, amazonConfig AmazonCognitoConfiguration, claimsParser ClaimsParser, session *sessions.CookieStore) *AmazonSupport {
	return &AmazonSupport{client, amazonConfig, claimsParser, session}
}

type ClaimsParser interface {
	ParseWithClaims(tokenString string, region string, claims jwt.Claims) (*jwt.Token, error)
}

type AmazonCognitoClaimsParser struct {
}

func (m AmazonCognitoClaimsParser) ParseWithClaims(tokenString string, _ string, claims jwt.Claims) (*jwt.Token, error) {
	log.Println("Enabling amazon cognito middleware.")

	// todo - currently unable to parse and verify the amazon elb token as the elb returns an un-parsable base64 token
	tokenString = strings.Replace(tokenString, "=", "", -1)

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return token, err
	}
	claims = token.Claims.(*AmazonCognitoClaims)
	return nil, nil
}

func (a *AmazonSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if access := r.Header["X-Amzn-Oidc-Data"]; access != nil && len(access) > 0 {
			claims := &AmazonCognitoClaims{}
			_, tokenErr := a.claimsParser.ParseWithClaims(access[0], a.amazonConfig.Region, claims)
			if tokenErr != nil {
				log.Printf("Oops, error parsing amazon cognito token claims. %v\n", tokenErr.Error())
			} else {
				session, _ := a.session.Get(r, "session")
				log.Println(fmt.Sprintf("Found amazon cognito authenticated user email %v", claims.Email))
				session.Values["principal"] = []string{claims.Email}
				session.Values["logout"] = fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/logout?client_id=%v&redirect_uri=%s&response_type=code",
					a.amazonConfig.Domain, a.amazonConfig.Region, a.amazonConfig.UserPoolClientId, a.amazonConfig.RedirectUrl)
				_ = session.Save(r, w)
			}
		}
		next.ServeHTTP(w, r)
	})
}
