PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

# Shell completion directories (XDG / standard locations)
ZSH_COMPLETION_DIR   ?= $(HOME)/.zsh/completions
BASH_COMPLETION_DIR  ?= $(HOME)/.local/share/bash-completion/completions
FISH_COMPLETION_DIR  ?= $(HOME)/.config/fish/completions

.PHONY: all build test smoke install install-completions uninstall uninstall-completions clean release release-snapshot

all: build

build:
	@echo "Building aim..."
	go build -ldflags="-s -w" -o aim ./cmd/aim

test:
	@echo "Running tests..."
	go test -v -race ./...

smoke: build
	@echo "Running smoke tests..."
	./test/smoke_test.sh

release-snapshot:
	@echo "Building local snapshot release with GoReleaser..."
	goreleaser release --snapshot --clean

release:
	@echo "Publishing release with GoReleaser..."
	goreleaser release --clean

# ─── Install ────────────────────────────────────────────────────────────────

install: build install-completions
	@echo "Installing aim to $(DESTDIR)$(BINDIR)..."
	@mkdir -p $(DESTDIR)$(BINDIR)
	@install -m 755 aim $(DESTDIR)$(BINDIR)/aim
	@echo "✓ aim installed to $(DESTDIR)$(BINDIR)/aim"
	@if ! echo "$$PATH" | grep -q "$(BINDIR)"; then \
		echo ""; \
		echo "  Note: add $(BINDIR) to your PATH:"; \
		echo "    export PATH=\"$(BINDIR):\$$PATH\""; \
	fi

install-completions: build
	@echo "Installing shell completions..."

	@# ── zsh ──────────────────────────────────────────────────────────────
	@mkdir -p $(DESTDIR)$(ZSH_COMPLETION_DIR)
	@./aim completion zsh > $(DESTDIR)$(ZSH_COMPLETION_DIR)/_aim
	@echo "  ✓ zsh  → $(DESTDIR)$(ZSH_COMPLETION_DIR)/_aim"
	@echo "    Add to ~/.zshrc if not already present:"
	@echo "      fpath=($(ZSH_COMPLETION_DIR) \$$fpath)"
	@echo "      autoload -Uz compinit && compinit"

	@# ── bash ─────────────────────────────────────────────────────────────
	@mkdir -p $(DESTDIR)$(BASH_COMPLETION_DIR)
	@./aim completion bash > $(DESTDIR)$(BASH_COMPLETION_DIR)/aim
	@echo "  ✓ bash → $(DESTDIR)$(BASH_COMPLETION_DIR)/aim"
	@echo "    Add to ~/.bashrc if not already present:"
	@echo "      source $(BASH_COMPLETION_DIR)/aim"

	@# ── fish ─────────────────────────────────────────────────────────────
	@mkdir -p $(DESTDIR)$(FISH_COMPLETION_DIR)
	@./aim completion fish > $(DESTDIR)$(FISH_COMPLETION_DIR)/aim.fish
	@echo "  ✓ fish → $(DESTDIR)$(FISH_COMPLETION_DIR)/aim.fish"
	@echo "    Fish picks this up automatically on next shell start."

# ─── Uninstall ───────────────────────────────────────────────────────────────

uninstall: uninstall-completions
	@echo "Uninstalling aim..."
	@rm -f $(DESTDIR)$(BINDIR)/aim
	@echo "✓ Removed $(DESTDIR)$(BINDIR)/aim"

uninstall-completions:
	@echo "Removing shell completions..."
	@rm -f $(DESTDIR)$(ZSH_COMPLETION_DIR)/_aim   && echo "  ✓ zsh  removed" || true
	@rm -f $(DESTDIR)$(BASH_COMPLETION_DIR)/aim    && echo "  ✓ bash removed" || true
	@rm -f $(DESTDIR)$(FISH_COMPLETION_DIR)/aim.fish && echo "  ✓ fish removed" || true

# ─── Clean ───────────────────────────────────────────────────────────────────

clean:
	@echo "Cleaning up..."
	@rm -f aim
