echo "\nHexa Orchestrator builder utility\n"

source ./.env

tag="hexaorchestrator"
test="N"
doPush="N"
platform=""
multi="N"
optString="srmhtdcpn:"
start="N"
reset="N"
while getopts ${optString} OPTION; do
  case "$OPTION" in
    t)
          test="Y"
          ;;
    n)
      tag=${OPTARG}
      echo "  .. using docker tag: $tag"
      ;;
    p)
      echo "  .. push to Docker Hub requested"
      doPush="Y"
      ;;
    c)
      echo "* Installing Hexa CLI & KeyTool"

      cat << EOF

  Executables are installed in the directory named by the GOBIN environment
  variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
  environment variable is not set. Executables in $GOROOT
  are installed in $GOROOT/bin or $GOTOOLDIR instead of $GOBIN.

  If environment variable HEXA_HOME is not set, HEXA configuration data is stored
  and loaded from ~/.hexa/config.json
EOF

      echo "* installing 'hexa' CLI"
      go install github.com/hexa-org/policy-mapper/cmd/hexa@latest

      echo "* installing 'hexaKey' tool"
      go install ./cmd/hexaKey

      exit
      ;;

    m)
      echo " .. multi platform build selected"
      multi="Y"
      ;;
    r)
      echo " ..reset environment requested"

      read -p "Are you sure? " -n 1 -r
      echo    # (optional) move to a new line
      if [[ $REPLY =~ ^[Yy]$ ]]
      then
          reset="Y"
      fi

      ;;
    s)
      echo "  ..startup requested"
      start="Y"
      ;;
    *)
      echo "Usage: ./build.sh -t -n <tag> "
      echo "  -c         Check and install the Hexa CLI from github.com/hexa-org/policy-mapper"
      echo "  -m         Build for multi-platform (requires docker with containerd configured)"
      echo "  -n <value> Builds the docker image with the specified tag name [hexaopa]"
      echo "  -p         Push the image to docker [default: not pushed]"
      echo "  -r         RESET all data to default (wipe keycloak, keys, and demo policy)"
      echo "  -s         Start Docker demo environment"
      echo "  -t         Performs build and test (default: build only)"

      exit 1
  esac
done

source ./.env

if [ "$test" = 'Y' ];then
    echo "* Building and running tests ..."
    source ./.env_development
    go build ./...
    go test ./...
    echo ""
fi

if [ "$reset" = 'Y' ];then
  echo "* Resetting environment"
  echo "  .. removing keys"
  rm -fv ./.certs/*
  echo "  .. removing demo policy"
  rm -dRfv ./deployments/hexaBundleServer/resources/bundles/bundle
  echo " .. removing docker containers"
  docker compose down
fi

echo "* Building containers"
echo "  .. building go linux executables for docker ..."

CGO_ENABLED=0 GOOS=linux go build -o ./hexaAdminUi  cmd/admin/admin.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaOrchestrator cmd/orchestrator/orchestrator.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaKeytool cmd/hexaKeytool/main.go

echo "  .. downloading latest chainguard platform image"
docker pull cgr.dev/chainguard/static:latest

echo "  .. building docker container image ($tag)..."

if [ "$multi" = 'Y' ];then
   echo "  .. performing multi platform build"
   docker build --platform=linux/amd64,linux/arm64 --tag "$tag" .
else
  echo "  .. building for local platform"
  docker build --tag "$tag" .
fi

if [ "$doPush" = 'Y' ];then
    echo "  .. pushing to docker ..."
    docker push $tag
fi

echo "  .. build complete."

if [ "$start" = 'Y' ];then
  echo "* Starting docker services"
  docker compose up -d
fi