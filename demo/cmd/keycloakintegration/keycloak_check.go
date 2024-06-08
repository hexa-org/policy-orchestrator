package main

import (
	"encoding/json"
	"fmt"

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

	token, err := handler.GetToken()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tokenBytes, _ := json.MarshalIndent(token, "", " ")
	fmt.Println(string(tokenBytes))
}
