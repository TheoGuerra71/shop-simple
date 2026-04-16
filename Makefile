.PHONY: build run dev test vet lint clean docker-up docker-down

# Binário de saída
BINARY=simple-shop

## build: Compila o binário
build:
	go build -o $(BINARY) ./cmd/api

## run: Compila e roda
run: build
	./$(BINARY)

## dev: Roda direto sem compilar binário
dev:
	go run ./cmd/api

## test: Roda os testes
test:
	go test ./... -v

## vet: Análise estática
vet:
	go vet ./...

## lint: Verifica formatação
lint:
	gofmt -l .

## clean: Remove o binário
clean:
	rm -f $(BINARY)

## docker-up: Sobe com Docker Compose
docker-up:
	docker compose up --build -d

## docker-down: Para os containers
docker-down:
	docker compose down

## help: Mostra comandos disponíveis
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
