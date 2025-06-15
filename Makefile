.PHONY: build clean run docker-build docker-run version help download_data data all test lint count-files verify build-verify build-gpmf

# 建置參數
BINARY_NAME=photo-sorter
DOCKER_IMAGE=photo-sorter
MAIN_PATH=cmd/photo-sorter/main.go
CONFIG_PATH=configs/config.yaml

# 從配置檔案讀取版本號（移除註解和引號）
VERSION = $(shell grep '^version:' $(CONFIG_PATH) | sed 's/version: *//' | sed 's/"//g' | sed 's/\#.*$$//' | xargs)
BUILD_TIME = $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse HEAD)

# 建置應用程式
build:
	go build -ldflags "-X photo-sorter/internal/pkg/version.Version=$(VERSION) \
		-X photo-sorter/internal/pkg/version.BuildTime=$(BUILD_TIME) \
		-X photo-sorter/internal/pkg/version.GitCommit=$(GIT_COMMIT)" \
		-o $(BINARY_NAME) $(MAIN_PATH)

# 清理建置檔案
clean:
	go clean
	rm -f $(BINARY_NAME)

# 執行應用程式
run: build
	./$(BINARY_NAME) -c $(CONFIG_PATH)

# 顯示版本資訊
version: build
	./$(BINARY_NAME) -version

# 建置 Docker 映像
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# 執行 Docker 容器
docker-run:
	docker run -v $(PWD):/app $(DOCKER_IMAGE):$(VERSION) -c $(CONFIG_PATH)

download_data:
	if [ ! -f ./vsizip/ne_10m_admin_1_states_provinces.zip ]; then curl -L -o ./vsizip/ne_10m_admin_1_states_provinces.zip https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip; fi

data: download_data
	rm -rf geodata/states.geojson
	unzip -o ./vsizip/ne_10m_admin_1_states_provinces.zip -d ./vsizip
	ogr2ogr -f GeoJSON -overwrite -makevalid -lco COORDINATE_PRECISION=6 \
	-sql "SELECT admin, name_en as name, adm0_a3 FROM ne_10m_admin_1_states_provinces" \
	geodata/states.geojson ./vsizip/ne_10m_admin_1_states_provinces.shp

data-sqlite:
	rm -rf geodata/states.sqlite
	unzip -o ./vsizip/ne_10m_admin_1_states_provinces.zip -d ./vsizip
	ogr2ogr -f SQLite -dsco SPATIALITE=YES -nlt PROMOTE_TO_MULTI geodata/states.sqlite ./vsizip/ne_10m_admin_1_states_provinces.shp


# 預設目標
all: build

# 測試
test:
	go test -v ./...

# 程式碼檢查
lint:
	golangci-lint run

# 計算檔案數量
count-files:
	@if [ -z "$(path)" ]; then \
		echo "錯誤：請提供目錄路徑"; \
		echo "使用方式：make count-files path=/path/to/directory"; \
		exit 1; \
	fi
	@chmod +x scripts/count_files.sh
	@./scripts/count_files.sh "$(path)"

# 驗證目錄
verify:
	@if [ -z "$(source)" ] || [ -z "$(target)" ]; then \
		echo "錯誤：請提供來源和目標目錄路徑"; \
		echo "使用方式：make verify source=/path/to/source target=/path/to/target"; \
		exit 1; \
	fi
	@chmod +x scripts/verify.sh
	@./scripts/verify.sh "$(source)" "$(target)"

verify-bin: build-verify
	./bin/verify -source "$(source)" -target "$(target)"

# 編譯 verify 工具
build-verify:
	go build -o bin/verify cmd/verify/main.go

# 定義變數
GPMF_DIR := third_party/gpmf-parser
GPMF_BUILD_DIR := $(GPMF_DIR)/build
TOOLS_BIN_DIR := tools/bin
GPMF_PARSER := gpmf-parser

build-gpmf:
	@echo "Building gpmf-parser..."
	@mkdir -p $(TOOLS_BIN_DIR)
	@cd $(GPMF_DIR) && \
		mkdir -p build && \
		cd build && \
		cmake .. && \
		make && \
		chmod +x $(GPMF_PARSER) && \
		cp $(GPMF_PARSER) ../../../$(TOOLS_BIN_DIR)/$(GPMF_PARSER)
	@rm -rf $(GPMF_BUILD_DIR)
	@echo "gpmf-parser build completed."

# 顯示幫助資訊
help:
	@echo "可用的目標："
	@echo "  build        - 建置應用程式"
	@echo "  clean        - 清理建置檔案"
	@echo "  run          - 建置並執行應用程式"
	@echo "  version      - 顯示版本資訊"
	@echo "  docker-build - 建置 Docker 映像"
	@echo "  docker-run   - 執行 Docker 容器"
	@echo "  count-files  - 計算目錄中的檔案數量"
	@echo "  verify       - 比對兩個目錄的檔案差異"
	@echo "  help         - 顯示此幫助資訊"
