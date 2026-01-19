BINARY_NAME := plantir
INSTALL_DIR := $(HOME)/local/bin
COMPLETION_DIR := $(HOME)/.zsh/completions

.PHONY: build install completion clean all

build:
	go build -o $(BINARY_NAME) .

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

completion: build
	@mkdir -p $(COMPLETION_DIR)
	./$(BINARY_NAME) completion zsh > $(COMPLETION_DIR)/_$(BINARY_NAME)
	@echo "Installed zsh completion to $(COMPLETION_DIR)/_$(BINARY_NAME)"
	@echo ""
	@echo "Add this to your ~/.zshrc if not already present:"
	@echo '  fpath=(~/.zsh/completions $$fpath)'
	@echo '  autoload -Uz compinit && compinit'

all: install completion
	@echo ""
	@echo "Done! Restart your shell or run: source ~/.zshrc"

clean:
	rm -f $(BINARY_NAME)
