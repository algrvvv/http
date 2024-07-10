BIN_DIR := bin
EXECUTABLE := http
SRC := cmd/http/main.go

install:
	go install $(SRC)

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(EXECUTABLE) $(SRC)

gr:
	go run $(SRC)

run: build
	./$(BIN_DIR)/$(EXECUTABLE)

clean:
	rm -rf $(BIN_DIR)/
