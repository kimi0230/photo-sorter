FROM golang:1.23-alpine

# 安裝 exiftool
RUN apk add --no-cache exiftool

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製原始碼
COPY . .

# 建置應用程式
RUN go build -o photo-sorter

# 設定預設命令
ENTRYPOINT ["./photo-sorter"] 
