APP_NAME := cabinete
SRC := main.go
INSTALL_DIR := /usr/local/bin

.PHONY: all build run clean install uninstall install-deps

all: build

build:
	go build -o $(APP_NAME) $(SRC)

run: build
	./$(APP_NAME) -d $(DIR)

clean:
	rm -f $(APP_NAME)

install-deps:
	go mod tidy
	go mod download

install: build
	@echo "Installing $(APP_NAME) to $(INSTALL_DIR)..."
	@sudo cp $(APP_NAME) $(INSTALL_DIR)
	@sudo chmod +x $(INSTALL_DIR)/$(APP_NAME)
	@echo "$(APP_NAME) has been installed to $(INSTALL_DIR)"

uninstall:
	@echo "Uninstalling $(APP_NAME) from $(INSTALL_DIR)..."
	@sudo rm -f $(INSTALL_DIR)/$(APP_NAME)
	@echo "$(APP_NAME) has been removed from $(INSTALL_DIR)"

