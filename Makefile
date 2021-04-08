BUILD_DIR = bin

UNAME_S := $(shell uname -s | tr A-Z a-z)

GO_PKGS := $(shell go list ./...)

NAME = cadence

build: build/$(UNAME_S) ## build binaries based on the OS
	@ln -s $(NAME).$(UNAME_S) $(BUILD_DIR)/$(NAME) || true

build/$(UNAME_S):
	@echo "$@"
	@rm -rf bin/*
	@GOOS=$(UNAME_S) GO111MODULE=on go build -o $(BUILD_DIR)/$(NAME).$(UNAME_S) github.com/ContextLogic/$(NAME)/cmd
