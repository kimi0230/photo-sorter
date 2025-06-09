.PHONY: build clean run docker-build docker-run version help download_data data

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


# 顯示幫助資訊
help:
	@echo "可用的目標："
	@echo "  build        - 建置應用程式"
	@echo "  clean        - 清理建置檔案"
	@echo "  run          - 建置並執行應用程式"
	@echo "  version      - 顯示版本資訊"
	@echo "  docker-build - 建置 Docker 映像"
	@echo "  docker-run   - 執行 Docker 容器"
	@echo "  help         - 顯示此幫助資訊"
