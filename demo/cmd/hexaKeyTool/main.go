/*
hexaKeyTool is a command line tool that can be used to generate a set of self-signed keys for use by
the hexaBundleServer, hexaOpa server, and the Hexa AuthZen server.

USAGE:

	hexaKeyTool -type=tls
	hexaKeyTool -type=jwt -action=init -dir=./certs

This will generate a CA cert/key pair and use that to sign Server cert/key pair
and Client cert/key pair.

Use these certs for tests such as websupport_test and orchestrator_test.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexa-org/policy-mapper/pkg/keysupport"
	"github.com/hexa-org/policy-mapper/pkg/tokensupport"
)

func doTlsKeys() {
	config := keysupport.GetKeyConfig()
	// get our ca and server certificate
	err := config.InitializeKeys()
	if err != nil {
		panic(err)
	}
}

func main() {

	pathHome := os.Getenv("HOME")
	keyPath := filepath.Join(pathHome, "./.certs")

	typeFlag := flag.String("type", "jwt", "one of tls|jwt")
	cmdFlag := flag.String("action", "token", "one of init|issue")
	dirFlag := flag.String("dir", keyPath, "filepath for storing keys")
	keyfileFlag := flag.String("keyfile", "", "Path to existing private key")
	scopeFlag := flag.String("scopes", "az", "az,bundle,root")
	mailFlag := flag.String("mail", "", "email address for user of token")
	helpFlag := flag.Bool("help", false, "To return help")

	flag.Parse()

	arg := flag.Arg(0)
	if (helpFlag != nil && *helpFlag) || strings.EqualFold("help", arg) {
		fmt.Println(`
Keytool generates certificates and tokens for use with the Hexa Bundle Server and AuthZen server

To generate TLS certificates for the bundle server use:
keytool -type=tls

To create a JWT certificate issuer use
keytool -type=jwt --action=init --dir=./certs`)
		return
	}

	if dirFlag != nil {
		os.Setenv(tokensupport.EnvTknKeyDirectory, *dirFlag)
	}

	switch strings.ToLower(*typeFlag) {
	case "tls":
		doTlsKeys()
	case "jwt":
		_, err := os.Stat(*dirFlag)
		if os.IsNotExist(err) {
			_ = os.Mkdir(*dirFlag, 0755)
		}
		keyFileName := filepath.Join(*dirFlag, tokensupport.DefTknPrivateKeyFile)

		switch strings.ToLower(*cmdFlag) {
		case "init":
			handler, err := tokensupport.GenerateIssuerKeys("authzen", false)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(fmt.Sprintf("Token public and private keys generated in %s", handler.KeyDir))
		case "issue":
			useKey := keyFileName
			if keyfileFlag != nil && *keyfileFlag != "" {
				useKey = *keyfileFlag
			}
			os.Setenv(tokensupport.EnvTknPrivateKeyFile, useKey)
			handler, err := tokensupport.LoadIssuer("authzen")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if scopeFlag == nil || *scopeFlag == "" {
				fmt.Println("Missing value for --scopes")
				return
			}
			scopes := strings.Split(strings.ToLower(*scopeFlag), ",")
			for _, scope := range scopes {
				switch scope {
				case tokensupport.ScopeAdmin, tokensupport.ScopeBundle, tokensupport.ScopeDecision:
					// ok
				default:
					fmt.Println(fmt.Printf("Invalid scope [%s] detected.", scope))
					return
				}
			}
			if mailFlag == nil || *mailFlag == "" {
				fmt.Println("An email address (-mail) is required for the user of the token")
				return
			}
			var tokenString string
			tokenString, err = handler.IssueToken(scopes, *mailFlag)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Bearer token issued:")
			fmt.Println(tokenString)
		default:
			fmt.Println("Select -action=init or -action=issue")
		}
	default:
		fmt.Println("Select -type=jwt or -type=tls")
	}
}
