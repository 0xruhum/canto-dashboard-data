run-catchup :; go run ./cmd/catchup/main.go

run-retry :; go run ./cmd/retry/main.go

anvil :; anvil --fork-url https://rpc.cantoarchive.com --no-mining --no-rate-limit --no-storage-caching --port 8546
