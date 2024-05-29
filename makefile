GO := go
YARN := yarn
VERSION := $(shell git describe --tags --abbrev=0)

GO_LDFLAGS := "-X 'github.com/dangjinghao/uclipboard/model.Version=$(VERSION)' -s -w" 

BUILD_DIR := build
FRONTEND_DIR := frontend-repo
FRONTEND_DIST := server/frontend/dist

SRCS := $(shell find . -type f -name "*.go")
FRONTEND_SRCS := $(shell find $(FRONTEND_DIR)/src $(FRONTEND_DIR)/public -type f )

TARGET := uclipboard

TMP_BUILD_CMD := $(GO) build -ldflags="-X 'github.com/dangjinghao/uclipboard/model.Version=$(shell git describe --tags --always --dirty)'" -o $(BUILD_DIR)/$(TARGET) . #ignore optimization for debug

LOG_LEVEL := info

build: $(BUILD_DIR)/$(TARGET)
all: bin docker-image
bin: $(SRCS) $(FRONTEND_DIST)/index.html
	@echo "uclipboard version: $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	@echo "multi-platform compiling..."
	@bash ./build_all.sh $(BUILD_DIR) $(TARGET) $(GO_LDFLAGS)
	@bash ./build_all.sh $(BUILD_DIR) $(TARGET) $(GO_LDFLAGS)

docker-image: bin
	@echo "building container"
	@docker build -t djh233/uclipboard:$(VERSION) .
	@docker build -t djh233/uclipboard:latest .
	
$(BUILD_DIR)/$(TARGET): $(SRCS) $(FRONTEND_DIST)/index.html
	@mkdir -p $(BUILD_DIR)
	@echo "building $(TARGET) without any optimization"
	@GOOS=linux GOARCH=amd64 $(TMP_BUILD_CMD)

$(FRONTEND_DIST)/index.html: $(FRONTEND_SRCS) 
	@echo "building frontend"
	@cd $(FRONTEND_DIR) && $(YARN) install && $(YARN) build
	@echo "moving frontend dist to server"
	cp -rT $(FRONTEND_DIR)/dist/ $(FRONTEND_DIST)/


build-frontend: $(FRONTEND_DIST)/index.html

clean:
	@rm -f $(BUILD_DIR)/*
	@rm -rf $(FRONTEND_DIST)/*
	
run: $(BUILD_DIR)/$(TARGET)
	@echo "run one local clinet and server on tmux" && sleep 1
	@tmux new-session -n run_uclipboard "$(SHELL) -c '$(BUILD_DIR)/$(TARGET) --mode server --log-level $(LOG_LEVEL); $(SHELL)'"  \
		\; split-window -h "$(SHELL) -c '$(BUILD_DIR)/$(TARGET) --mode client --log-level $(LOG_LEVEL); $(SHELL)'" 

run-server: $(BUILD_DIR)/$(TARGET)
	@echo "run local server"
	@$(BUILD_DIR)/$(TARGET) --mode server --log-level $(LOG_LEVEL)

.PHONY: clean all run build-frontend run-server build docker-image bin
