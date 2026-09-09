.PHONY: build install-cli install-app clean run scan test gui app

BINARY=kclean
APP_NAME=K-Cleaner.app
INSTALL_PATH=/usr/local/bin

build:
	go build -ldflags="-s -w" -o $(BINARY) .

app: build
	@bash scripts/build-app.sh

install-cli: build
	install -m 755 $(BINARY) $(INSTALL_PATH)/$(BINARY)

install-app: app
	cp -R $(APP_NAME) /Applications/

clean:
	rm -f $(BINARY)
	rm -rf $(APP_NAME) build/
	go clean

run: build
	./$(BINARY) scan

gui: build
	./$(BINARY) gui

scan: build
	./$(BINARY) scan --verbose

test:
	go test ./...

dry-run: build
	./$(BINARY) clean --all --dry-run --yes
