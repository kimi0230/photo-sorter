.PHONY: build clean run docker-build docker-run

# 建置參數
BINARY_NAME=photo-sorter
DOCKER_IMAGE=photo-sorter
MAIN_PATH=cmd/photo-sorter/main.go

# 建置應用程式
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# 清理建置檔案
clean:
	go clean
	rm -f $(BINARY_NAME)

# 執行應用程式
run: build
	./$(BINARY_NAME) -config config/config.yaml

# 建置 Docker 映像
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# 執行 Docker 容器
docker-run:
	docker run -v $(PWD):/app $(DOCKER_IMAGE) -config config/config.yaml
