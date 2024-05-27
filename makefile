GO := go

FRONTEND_REPO_URL := https://github.com/uclipboard/frontend

PROJECT_DIR := .
BUILD_DIR := $(PROJECT_DIR)/
FRONTEND_DIR := $(PROJECT_DIR)/tmp/frontend

SRCS := $(shell find . -type f -name "*.go")
WEB_DIST := $(shell find $(PROJECT_DIR)/server/frontend/dist -type f )

TARGET := uclipboard
TARGET_WIN := $(TARGET).exe

BUILD_CMD := $(GO) build -ldflags="-s -w"  -o $(BUILD_DIR)/$(TARGET) $(PROJECT_DIR)
BUILD_WIN_CMD := $(GO) build -ldflags="-s -w"  -o $(BUILD_DIR)/$(TARGET_WIN) $(PROJECT_DIR)

GOOS_WIN=windows
GOOS_LINUX=linux
GOARCH_AMD64=amd64

LOG_LEVEL := info

all: $(TARGET) $(TARGET_WIN)

$(TARGET): $(SRCS) $(WEB_DIST)
	@mkdir -p $(BUILD_DIR)
	@echo "building $(TARGET)"
	@GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_AMD64)
	@$(BUILD_CMD)

build-frontend-image:
	@echo "building frontend image"
	@echo "TODO: implement"
build-frontend: | $(FRONTEND_DIR) build-frontend-image
	@cd $(FRONTEND_DIR) && git pull
	@echo "TODO: implement"
	
$(FRONTEND_DIR):
	@echo "cloning frontend"
	@git clone $(FRONTEND_REPO_URL) $(FRONTEND_DIR)
	
$(TARGET_WIN): $(SRCS) $(WEB_DIST)
	@mkdir -p $(BUILD_DIR)
	@echo "cross-building $(TARGET_WIN)"
	@GOOS=$(GOOS_WIN) GOARCH=$(GOARCH_AMD64)
	@$(BUILD_WIN_CMD)

clean:
	@rm -f $(BUILD_DIR)/$(TARGET) $(BUILD_DIR)/$(TARGET_WIN)
	@rm -rf $(PROJECT_DIR)/tmp/frontend
	@rm -f  $(PROJECT_DIR)/server/frontend/dist/*
	
run: $(TARGET)
	@echo "run local clinet and server on tmux"
	@sleep 1
	@tmux new-session -n run_uclipboard "$(SHELL) -c '$(BUILD_DIR)/$(TARGET) --mode server --log-level $(LOG_LEVEL); $(SHELL)'"  \
		\; split-window -h "$(SHELL) -c '$(BUILD_DIR)/$(TARGET) --mode client --log-level $(LOG_LEVEL); $(SHELL)'" 
run-backend: $(TARGET)
	@echo "run local server"
	@$(BUILD_DIR)/$(TARGET) --mode server --log-level $(LOG_LEVEL)



.PHONY: clean all run build-frontend-image build-frontend run-backend
