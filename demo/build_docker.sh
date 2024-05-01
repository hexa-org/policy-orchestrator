echo "This builds a docker image for local execution"

echo "  building Orchestrator executables ..."
CGO_ENABLED=0 GOOS=linux go build -o ./hexaAdminUi  cmd/admin/admin.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaDemo cmd/demo/demo.go
CGO_ENABLED=0 GOOS=linux go build -o ./hexaOrchestrator cmd/orchestrator/orchestrator.go

echo "  building docker container image..."
docker build --tag hexaorchestrator .

echo "  Build complete. Execute using docker compose"