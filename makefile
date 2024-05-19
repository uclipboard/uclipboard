GO := go

PROJECT_DIR := .
BUILD_DIR := .

SRCS := $(shell find . -type f -name "*.go")

TARGET := uclipboard
TARGET_WIN := $(TARGET_WIN)

BUILD_CMD := $(GO) build -ldflags="-s -w"  -o $(BUILD_DIR)/$(TARGET) $(PROJECT_DIR)
BUILD_WIN_CMD := $(GO) build -ldflags="-s -w"  -o $(BUILD_DIR)/$(TARGET_WIN) $(PROJECT_DIR)

GOOS_WIN=windows
GOOS_LINUX=linux
GOARCH_AMD64=amd64

LOG_LEVEL := info

all: $(TARGET) $(TARGET_WIN)

$(TARGET): $(SRCS)
	@mkdir -p $(BUILD_DIR)
	@echo "building $(TARGET)"
	@GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_AMD64)
	@$(BUILD_CMD)

$(TARGET_WIN): $(SRCS)
	@mkdir -p $(BUILD_DIR)
	@echo "cross-building $(TARGET_WIN)"
	@GOOS=$(GOOS_WIN) GOARCH=$(GOARCH_AMD64)
	@$(BUILD_WIN_CMD)
clean:
	@rm -f $(BUILD_DIR)/$(TARGET) $(BUILD_DIR)/$(TARGET_WIN)

run: $(TARGET)
	@echo "run local clinet and server on tmux"
	@tmux new-session -n run_uclipboard "$(BUILD_DIR)/$(TARGET) --mode server --log-level $(LOG_LEVEL)"  \
		\; split-window -h "$(BUILD_DIR)/$(TARGET) --mode client --log-level $(LOG_LEVEL)" 

.PHONY: clean all run
