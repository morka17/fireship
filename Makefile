
client:
	@go build -o bin/client game_client/main.go


server:
	@go build -o bin/server game_server/main.go