package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oauth2support"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	config := clientcredentials.Config{
		ClientID:     "hexaclient",
		ClientSecret: "uuXVzfbqH635Ob0oTON1uboONUqasmTt",
		TokenURL:     "http://localhost:8080/realms/Hexa-Orchestrator-Realm/protocol/openid-connect/token",
		AuthStyle:    oauth2.AuthStyle(oauth2.AuthStyleAutoDetect),
	}

	handler := oauth2support.NewJwtClientHandlerWithConfig(&config)

	tokenSet, err := handler.GetToken()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tokenBytes, _ := json.MarshalIndent(tokenSet, "", " ")
	fmt.Println(string(tokenBytes))

	ats := tokenSet.AccessToken

	fmt.Println()
	fmt.Println("Access Token...")
	jwkKeyfunc, err := keyfunc.NewDefaultCtx(context.Background(), []string{"http://localhost:8080/realms/Hexa-Orchestrator-Realm/protocol/openid-connect/certs"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	at, err := jwt.Parse(ats, jwkKeyfunc.Keyfunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	atBytes, _ := json.MarshalIndent(at, "", " ")
	fmt.Println(string(atBytes))
}
