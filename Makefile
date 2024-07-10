BIN_DIR := bin
EXECUTABLE := http
SRC := cmd/http/main.go

install:
	@go install $(SRC)

build:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(EXECUTABLE) $(SRC)
	@echo "билд заверешен"

gr:
	@go run $(SRC)

run: build
	@./$(BIN_DIR)/$(EXECUTABLE)

clean:
	@rm -rf $(BIN_DIR)/
	@echo "папка билда очищена"
