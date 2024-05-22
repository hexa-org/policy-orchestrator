
echo "Building and testing Policy-Orchestrator."
echo "  if running manually, see demo/build.sh"
# This is called by CodeQL Github Action

cd ./demo
source ./.env_development
source ./build.sh -b