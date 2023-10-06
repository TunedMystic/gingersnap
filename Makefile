APP="$$(basename -- $$(PWD))"


# -------------------------------------------------------------------
# App-related commands
# -------------------------------------------------------------------

## @(app) - Run the Go app  --watch     ⭐️
run: bin/watchexec app/assets/bin/cwebp
	@echo "✨📦✨ Running the app server\n"
	@./bin/watchexec -r -e go,css,js,html,md,json "go run ./cmd/server/"


## @(app) - Run tailwindcss --watch     ⭐️
css: bin/tailwind
	$(MAKE) tailwind args=--watch


## @(app) - Build the app binary
build: clean app/assets/bin/cwebp
	@echo "✨📦✨ Building the app binary\n"
	@go build -ldflags="-s -w -X 'main.BuildHash=$$(git rev-parse --short=10 HEAD)' -X 'main.BuildDate=$$(date)'" -o bin/gingersnap ./cmd/cli/


## @(app) - Build the app binary and copy it to $GOPATH
install: build
	@echo "✨📦✨ Copying to \$$GOPATH\n"
	@cp bin/gingersnap $$GOPATH/bin/


## @(app) - Remove temp files and dirs
clean:
	@rm -f coverage.out
	@rm -f db.sqlite-*
	@go clean -testcache
	@find . -name '.DS_Store' -type f -delete
	@bash -c "mkdir -p bin && cd bin && find . ! -name 'watchexec' ! -name 'cwebp' ! -name 'tailwind' -type f -exec rm -f {} +"
	@rm -f gingersnap.json
	@rm -rf posts
	@rm -rf media
	@rm -f $$GOPATH/bin/gingersnap


tailwind: bin/tailwind
	@echo "✨📦✨ Running tailwind\n"
	@bash -c "./bin/tailwind --input ./tailwind.input.css --output ./assets/css/styles.css --minify $(args)"


bin/tailwind:
	@echo "✨📦✨ Downloading tailwindcss binary\n"
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mkdir -p bin
	mv tailwindcss-macos-arm64 ./bin/tailwind
	@echo ""


bin/watchexec:
	@echo "✨📦✨ Downloading watchexec binary\n"
	curl -sL https://github.com/watchexec/watchexec/releases/download/v1.23.0/watchexec-1.23.0-x86_64-apple-darwin.tar.xz | tar -xz
	mkdir -p bin
	mv ./watchexec-1.23.0-x86_64-apple-darwin/watchexec ./bin/watchexec
	rm -rf watchexec-1.23.0-x86_64-apple-darwin
	@echo ""


app/assets/bin/cwebp:
	@echo "✨📦✨ Downloading cwebp binary\n"
	curl -sL https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.1-mac-x86-64.tar.gz | tar -xz
	mkdir -p assets/bin
	mv ./libwebp-1.3.1-mac-x86-64/bin/cwebp ./app/assets/bin/cwebp
	rm -rf libwebp-1.3.1-mac-x86-64
	@echo ""



# -------------------------------------------------------------------
# Self-documenting Makefile targets - https://bit.ly/32lE64t
# -------------------------------------------------------------------

.DEFAULT_GOAL := help

help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ \
		{ \
			helpMessage = match(lastLine, /^## (.*)/); \
			if (helpMessage) { \
				helpCommand = substr($$1, 0, index($$1, ":")-1); \
				helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
				helpGroup = match(helpMessage, /^@([^ ]*)/); \
				if (helpGroup) { \
					helpGroup = substr(helpMessage, RSTART + 1, index(helpMessage, " ")-2); \
					helpMessage = substr(helpMessage, index(helpMessage, " ")+1); \
				} \
				printf "%s|  %-20s %s\n", \
					helpGroup, helpCommand, helpMessage; \
			} \
		} \
		{ lastLine = $$0 }' \
		$(MAKEFILE_LIST) \
	| sort -t'|' -sk1,1 \
	| awk -F '|' ' \
			{ \
			cat = $$1; \
			if (cat != lastCat || lastCat == "") { \
				if ( cat == "0" ) { \
					print "\nTargets:" \
				} else { \
					gsub("_", " ", cat); \
					printf "\n%s\n", cat; \
				} \
			} \
			print $$2 \
		} \
		{ lastCat = $$1 }'
	@echo ""
