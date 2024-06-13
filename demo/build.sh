echo "\nHexa Orchestrator builder utility\n"

tag="hexaorchestrator"
test="N"
doPush="N"
platform=""
multi="N"
optString="mhtdcpn:"
while getopts ${optString} OPTION; do
  case "$OPTION" in
    t)
          test="Y"
          ;;
        n)
          tag=${OPTARG}
          echo "  ..using docker tag: $tag"
          ;;
        p)
          echo "  ..push to Docker Hub requested"
          doPush="Y"
          ;;
        c)
          echo "* Installing Hexa CLI"
          if ! command -v hexa &> /dev/null
          then
              go install github.com/hexa-org/policy-mapper/cmd/hexa@latest
              exit 1
          fi
          hexa help
          exit
          ;;

        m)
          echo " ..multi platform build selected"
          multi="Y"
          ;;
    *)
      echo "Usage: ./build.sh -t -n <tag> "
      echo "  -t         Performs build and test (default: build only)"
      echo "  -m         Build for multi-platform (requires docker with containerd configured)"
      echo "  -n <value> Builds the docker image with the specified tag name [hexaopa]"
      echo "  -p         Push the image to docker [default: not pushed]"
      echo "  -c         Check and install the Hexa CLI from github.com/hexa-org/policy-mapper"
          exit 1
  esac
done

echo ""

if [ "$test" = 'Y' ];then
    echo "* Building and running tests ..."
    source ./.env_development
    go build ./...
    go test ./...
    echo ""
fi

echo "* building go linux executables for docker ..."

CGO_ENABLED=0 GOOS=linux go build -o ./hexaAdminUi  cmd/admin/admin.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaOrchestrator cmd/orchestrator/orchestrator.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaKeytool cmd/hexaKeytool/main.go

echo "* building docker container image ($tag)..."
echo "  - downloading latest chainguard platform image"
docker pull cgr.dev/chainguard/static:latest

if [ "$multi" = 'Y' ];then
   echo "  - performing multi platform build"
   docker build --platform=linux/amd64,linux/arm64 --tag "$tag" .
else
  echo "  - building for local platform"
  docker build --tag "$tag" .
fi

if [ "$doPush" = 'Y' ];then
    echo "  pushing to docker ..."
    docker push $tag
fi

echo "Build complete. Execute using docker compose"