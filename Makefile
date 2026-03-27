APP_NAME := Copilot CLI Session Monitor
BINARY_NAME := copilot-monitor
BUNDLE_NAME := Copilot CLI Session Monitor.app
INSTALL_DIR := /Applications

.PHONY: build bundle install clean

## build: Compile the Go binary
build:
	CGO_ENABLED=1 go build -o $(BINARY_NAME) .

## bundle: Build and package as a macOS .app
bundle: build
	rm -rf $(BUNDLE_NAME)
	mkdir -p "$(BUNDLE_NAME)/Contents/MacOS"
	mkdir -p "$(BUNDLE_NAME)/Contents/Resources"
	cp $(BINARY_NAME) "$(BUNDLE_NAME)/Contents/MacOS/"
	cp bundle/Info.plist "$(BUNDLE_NAME)/Contents/"
	cp assets/AppIcon.icns "$(BUNDLE_NAME)/Contents/Resources/"

## install: Install the .app to /Applications
install: bundle
	rm -rf "$(INSTALL_DIR)/$(BUNDLE_NAME)"
	cp -r "$(BUNDLE_NAME)" "$(INSTALL_DIR)/"
	@echo "Installed to $(INSTALL_DIR)/$(BUNDLE_NAME)"

## uninstall: Remove the .app from /Applications
uninstall:
	rm -rf "$(INSTALL_DIR)/$(BUNDLE_NAME)"
	@echo "Removed $(INSTALL_DIR)/$(BUNDLE_NAME)"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUNDLE_NAME)

## run: Build and run the binary directly (for development)
run: build
	./$(BINARY_NAME)
