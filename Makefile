# -------- configurable --------
BINARY ?= dnsctl
PREFIX ?= $(HOME)/go/bin
COMPLETIONS_DIR ?= $(HOME)/.oh-my-zsh/custom/completions
GO ?= go

# -------- derived --------
BIN_PATH := $(PREFIX)/$(BINARY)
# -------- targets --------

.PHONY: all build install completion release clean

all: build

## Build the binary
build:
	$(GO) build -o $(BINARY)

## Install binary + shell completions
install: build
	@echo "Installing $(BINARY) to $(PREFIX)"
	@mkdir -p $(PREFIX)
	@install -m 0755 $(BINARY) $(BIN_PATH)

	@$(MAKE) completion

## Install shell completion for current shell
completion:
	@echo "Installing shell completion"
	@SHELL_NAME="$$(basename "$$SHELL")"; \
	if [ "$$SHELL_NAME" = "bash" ]; then \
		RC_FILE="$(HOME)/.bashrc"; \
		LINE='source <($(BINARY) completion bash)'; \
	elif [ "$$SHELL_NAME" = "zsh" ]; then \
		RC_FILE="$(HOME)/.zshrc"; \
		LINE='source <($(BINARY) completion zsh)'; \
	else \
		echo "⚠ Unsupported shell: $$SHELL_NAME"; \
		echo "  Manually add one of the following:"; \
		echo "    source <($(BINARY) completion bash)"; \
		echo "    source <($(BINARY) completion zsh)"; \
		exit 0; \
	fi; \
	\
	touch "$$RC_FILE"; \
	if grep -Fq "$$LINE" "$$RC_FILE"; then \
		echo "✓ Completion already enabled in $$RC_FILE"; \
	else \
		echo "$$LINE" >> "$$RC_FILE"; \
		echo "✓ Added completion to $$RC_FILE"; \
	fi; \
	\
	echo ""; \
	echo "Restart your shell or run:"; \
	echo "  source $$RC_FILE"

## Build release binary (no completions, clean env)
release:
	@echo "Building release binary"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		$(GO) build -ldflags="-s -w" -o $(BINARY)

## Clean artifacts
clean:
	rm -f $(BINARY)
