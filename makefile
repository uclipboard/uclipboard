GO := go
YARN := yarn
VERSION := $(shell git describe --tags --abbrev=0)

GO_LDFLAGS := "-X 'github.com/uclipboard/uclipboard/model.Version=$(VERSION)' -s -w" 

BUILD_DIR := $(PWD)/build
FRONTEND_DIR := $(PWD)/frontend-repo
FRONTEND_DIST := $(PWD)/server/frontend/dist
WATCHER := $(PWD)/script/watcher.sh
BUILD_ALL := $(PWD)/script/build_all.sh

SRCS := $(shell find $(PWD) -type f -name "*.go")
FRONTEND_SRCS := $(shell find $(FRONTEND_DIR)/src $(FRONTEND_DIR)/public -type f )

TARGET := uclipboard

DEBUG_BUILD := $(GO) build -ldflags="-X 'github.com/uclipboard/uclipboard/model.Version=$(shell git describe --tags --always --dirty)'" -o $(BUILD_DIR)/$(TARGET) . #ignore optimization for debug

LOG_LEVEL := info

OTHER_ARGS := 

build: $(BUILD_DIR)/$(TARGET)
all: bin docker-image

bin: $(SRCS) $(FRONTEND_DIST)/index.html
	@echo "uclipboard version: $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	@echo "multi-platform compiling..."
	@$(BUILD_ALL) $(BUILD_DIR) $(TARGET) $(GO_LDFLAGS)

docker-image: bin
	@echo "building container"
	@docker build -t djh233/uclipboard:$(VERSION) .
	@docker build -t djh233/uclipboard:latest .

build-target-noweb: $(SRCS)
	@mkdir -p $(BUILD_DIR)
	@echo "building $(TARGET) without any optimization"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(DEBUG_BUILD)
	@echo "building completed"

$(BUILD_DIR)/$(TARGET): $(FRONTEND_DIST)/index.html build-target-noweb

$(FRONTEND_DIST)/index.html: $(FRONTEND_SRCS) 
	@echo "building frontend"
	@cd $(FRONTEND_DIR) && $(YARN) install && $(YARN) build
	@echo "moving frontend dist to server"
	@cp -rT $(FRONTEND_DIR)/dist/ $(FRONTEND_DIST)/


watch-build:
	@$(WATCHER) "build" "$(SRCS) $(FRONTEND_SRCS)" "$(YARN)" "$(LOG_LEVEL)" "$(OTHER_ARGS)"

watch-dev:
	@$(WATCHER) "dev" "$(SRCS)" "$(YARN)" "$(LOG_LEVEL)" "--test ctf" $(BUILD_DIR)/$(TARGET) #yarn dev will watch those FRONTEND_SRCS
	

dev-frontend:
	@echo "building frontend in dev mode"
	@cd $(FRONTEND_DIR) &&$(YARN) install &&$(YARN) dev --host


test:
	
clean:
	@rm -f $(BUILD_DIR)/uclipboard*
	@rm -rf $(FRONTEND_DIST)/*
	
run-client: $(BUILD_DIR)/$(TARGET)
	@echo "run local client"
	@cd $(BUILD_DIR) && ./$(TARGET) --mode client --log-level $(LOG_LEVEL)

run-server: $(BUILD_DIR)/$(TARGET)
	@echo "run local server"
	@cd $(BUILD_DIR) && ./$(TARGET) --mode server --log-level $(LOG_LEVEL) $(OTHER_ARGS)

run-server-noweb: $(SRCS)
	@make build-target-noweb
	@echo "run local server"
	@cd $(BUILD_DIR) && ./$(TARGET) --mode server --log-level $(LOG_LEVEL) $(OTHER_ARGS)

run-client-noweb: $(SRCS)
	@make build-target-noweb
	@echo "run local client"
	@cd $(BUILD_DIR) && ./$(TARGET) --mode client --log-level $(LOG_LEVEL) $(OTHER_ARGS)


.PHONY: clean all run build-frontend run-server build docker-image bin run-watch test dev-frontend build-target-noweb
