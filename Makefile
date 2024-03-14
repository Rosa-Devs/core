


api:
	nodemon --exec go run ./cmd/chat --ext go --signal SIGTERM

chat:
	go run ./cmd/chat
