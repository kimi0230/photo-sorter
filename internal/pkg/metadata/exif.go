package metadata

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
)

type ExifData struct {
	CreateDate      string `json:"CreateDate"`
	MediaCreateDate string `json:"MediaCreateDate"`
	DateTimeCreated string `json:"DateTimeCreated"`
	FileModifyDate  string `json:"FileModifyDate"`
	Model           string `json:"Model"`
	Encoder         string `json:"Encoder"`
	Description     string `json:"Description"`
	GPSLatitude     string `json:"GPSLatitude"`
	GPSLongitude    string `json:"GPSLongitude"`
}

//go:generate mockgen -destination=exif_mock.go -package=metadata . ExifReader
type ExifReader interface {
	GetExifData(ctx context.Context, path string) (*ExifData, error)
	Close() error
}

type LegacyExifReader struct{}

func NewLegacyExifReader() ExifReader {
	return &LegacyExifReader{}
}

func (r *LegacyExifReader) GetExifData(ctx context.Context, path string) (*ExifData, error) {
	return GetExifData(ctx, path)
}

func (r *LegacyExifReader) Close() error {
	return nil
}

type ExiftoolClient struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
	mu     sync.Mutex
}

// Warnf is a hook for reporting non-fatal warnings to the caller.
// By default it does nothing and can be overridden by the CLI/UI layer.
var Warnf = func(format string, args ...any) {}

type DateResolver interface {
	Resolve(exif *ExifData, cfg *config.Config) (string, error)
}

type DeviceResolver interface {
	Resolve(exif *ExifData) string
}

type GeoResolver interface {
	Resolve(baseDate string, exif *ExifData, cfg *config.Config) (string, *geocoding.CountryCity, error)
}

var exifDateFormats = []string{
	"2006:01:02 15:04:05",       // Standard EXIF format.
	"2006:01:02 15:04:05-07:00", // EXIF with timezone.
	"2006:01:02 15:04:05+07:00", // EXIF with timezone.
	"2006-01-02 15:04:05",       // ISO format.
	"2006-01-02 15:04:05-07:00", // ISO with timezone.
	"2006-01-02 15:04:05+07:00", // ISO with timezone.
}

// ParseGPSString 將 GPS 字串轉換為浮點數
// 格式範例: "22 deg 41' 58.80\" N"
func ParseGPSString(gpsStr string) (float64, error) {
	if gpsStr == "" {
		return 0, nil
	}

	// 移除引號和空格
	gpsStr = strings.Trim(gpsStr, "\"")
	gpsStr = strings.TrimSpace(gpsStr)

	// 分割字串
	parts := strings.Fields(gpsStr)
	if len(parts) < 4 {
		return 0, fmt.Errorf("invalid GPS format: %s", gpsStr)
	}

	// 解析度數
	degrees, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse degrees: %v", err)
	}

	// 解析分數
	minutes, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "'"), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse minutes: %v", err)
	}

	// 解析秒數
	seconds, err := strconv.ParseFloat(strings.TrimSuffix(parts[3], "\""), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse seconds: %v", err)
	}

	// 計算十進位度數
	decimal := degrees + minutes/60 + seconds/3600

	// 檢查方向（N/S 或 E/W）
	if len(parts) > 4 {
		direction := parts[4]
		if direction == "S" || direction == "W" {
			decimal = -decimal
		}
	}

	return decimal, nil
}

// isValidDate 檢查日期是否有效
// 將空字串和無效日期（如 0000:00:00 00:00:00）視為無效
func isValidDate(date string) bool {
	if date == "" {
		return false
	}
	// 檢查是否為無效日期格式
	invalidDates := []string{
		"0000:00:00 00:00:00",
		"0000-00-00 00:00:00",
		"0000:00:00",
		"0000-00-00",
	}
	for _, invalid := range invalidDates {
		if date == invalid {
			return false
		}
	}
	return true
}

type defaultDateResolver struct{}

func (r *defaultDateResolver) Resolve(exif *ExifData, cfg *config.Config) (string, error) {
	// 取得日期（優先順序：CreateDate > DateTimeCreated > MediaCreateDate > FileModifyDate）
	date := ""
	if isValidDate(exif.CreateDate) {
		date = exif.CreateDate
	} else if isValidDate(exif.DateTimeCreated) {
		date = exif.DateTimeCreated
	} else if isValidDate(exif.MediaCreateDate) {
		date = exif.MediaCreateDate
	} else if isValidDate(exif.FileModifyDate) {
		date = exif.FileModifyDate
	}

	if date == "" {
		return "unknown_date", nil
	}

	// 解析日期字串並使用設定檔中的格式
	// 嘗試多種日期格式
	var t time.Time
	var err error
	for _, format := range exifDateFormats {
		t, err = time.Parse(format, date)
		if err == nil {
			return t.Format(cfg.DateFormat), nil
		}
	}
	return "unknown_date", nil
}

type defaultDeviceResolver struct{}

func (r *defaultDeviceResolver) Resolve(exif *ExifData) string {
	// 取得裝置名稱
	// 優先使用 Model，如果沒有 Model 則使用 Encoder，最後才檢查 Description 是否為 "Screenshot"
	var device string
	if exif.Model != "" {
		device = getDeviceName(exif.Model)
	} else if exif.Encoder != "" {
		device = getDeviceName(exif.Encoder)
	} else {
		device = "unknown_device"
	}
	// 如果沒有裝置名稱，且 Description 是 "Screenshot"，則使用 "Screenshot"
	if device == "unknown_device" && strings.TrimSpace(exif.Description) == "Screenshot" {
		device = "Screenshot"
	}
	return device
}

func GetExifData(ctx context.Context, path string) (*ExifData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "exiftool", "-json", "-CreateDate", "-MediaCreateDate", "-DateTimeCreated", "-FileModifyDate", "-Model", "-Encoder", "-Description", "-GPSLatitude", "-GPSLongitude", "-ee", path)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("exiftool execution failed: %v", err)
	}

	var data []ExifData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse exiftool output: %v", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no file metadata found")
	}

	// 記錄執行時間
	executionTime := time.Since(startTime)
	if executionTime > 3*time.Second {
		Warnf("exiftool took %.2f seconds for %s", executionTime.Seconds(), path)
	}

	return &data[0], nil
}

func NewExiftoolClient() (*ExiftoolClient, error) {
	cmd := exec.Command("exiftool", "-stay_open", "True", "-@", "-")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to setup exiftool stdout: %v", err)
	}
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to setup exiftool stdin: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start exiftool: %v", err)
	}

	return &ExiftoolClient{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdinPipe),
		stdout: bufio.NewReader(stdoutPipe),
	}, nil
}

func (c *ExiftoolClient) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Fprintln(c.stdin, "-stay_open")
	fmt.Fprintln(c.stdin, "False")
	_ = c.stdin.Flush()

	return c.cmd.Wait()
}

func (c *ExiftoolClient) GetExifData(ctx context.Context, path string) (*ExifData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil {
		return GetExifData(ctx, path)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	args := []string{
		"-json",
		"-CreateDate",
		"-MediaCreateDate",
		"-DateTimeCreated",
		"-FileModifyDate",
		"-Model",
		"-Encoder",
		"-Description",
		"-GPSLatitude",
		"-GPSLongitude",
		"-ee",
		path,
		"-execute",
	}

	for _, arg := range args {
		if _, err := fmt.Fprintln(c.stdin, arg); err != nil {
			return nil, fmt.Errorf("failed to write exiftool args: %v", err)
		}
	}
	if err := c.stdin.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush exiftool stdin: %v", err)
	}

	type readResult struct {
		data []byte
		err  error
	}
	resultCh := make(chan readResult, 1)
	go func() {
		var buf bytes.Buffer
		for {
			line, err := c.stdout.ReadString('\n')
			if err != nil {
				resultCh <- readResult{err: fmt.Errorf("failed to read exiftool output: %v", err)}
				return
			}
			if strings.TrimSpace(line) == "{ready}" {
				resultCh <- readResult{data: buf.Bytes()}
				return
			}
			buf.WriteString(line)
		}
	}()

	var output []byte
	select {
	case <-ctx.Done():
		if c.cmd != nil && c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, result.err
		}
		output = result.data
	}

	var data []ExifData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse exiftool output: %v", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no file metadata found")
	}

	return &data[0], nil
}

// sanitizeDeviceName 清理並標準化裝置名稱
func sanitizeDeviceName(name string) string {
	// 移除所有非英文字母、數字和底線的字元
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, name)
}

// getDeviceName 取得標準化的裝置名稱
func getDeviceName(model string) string {
	if model == "" {
		return "unknown_device"
	}

	// 處理 DJI 機型名稱
	if deviceDJI, ok := GetDJIModelFriendlyName(model); ok {
		return deviceDJI
	}

	// 處理一般裝置名稱
	device := strings.ReplaceAll(model, " ", "_")
	return sanitizeDeviceName(device)
}

func GetTargetPath(path string, exif *ExifData, cfg *config.Config) (string, *geocoding.CountryCity, error) {
	return GetTargetPathWithGeoResolver(path, exif, cfg, NewDefaultGeoResolver(cfg))
}

func GetTargetPathWithGeoResolver(path string, exif *ExifData, cfg *config.Config, geoResolver GeoResolver) (string, *geocoding.CountryCity, error) {
	if geoResolver == nil {
		geoResolver = NewDefaultGeoResolver(cfg)
	}
	return GetTargetPathWithResolvers(
		path,
		exif,
		cfg,
		&defaultDateResolver{},
		&defaultDeviceResolver{},
		geoResolver,
	)
}

func GetTargetPathWithResolvers(
	path string,
	exif *ExifData,
	cfg *config.Config,
	dateResolver DateResolver,
	deviceResolver DeviceResolver,
	geoResolver GeoResolver,
) (string, *geocoding.CountryCity, error) {
	// 解析日期
	date, err := dateResolver.Resolve(exif, cfg)
	if err != nil {
		return "", nil, err
	}

	// 解析地理位置
	date, location, err := geoResolver.Resolve(date, exif, cfg)
	if err != nil {
		return "", nil, err
	}

	device := deviceResolver.Resolve(exif)

	// 建立目標路徑
	targetDir := filepath.Join(cfg.DstDir, date, device)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create target directory: %v", err)
	}

	// 處理檔案名稱衝突
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	targetPath := filepath.Join(targetDir, baseName)

	counter := 1
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		counter++
	}

	return targetPath, location, nil
}
