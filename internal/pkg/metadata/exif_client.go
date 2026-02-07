package metadata

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

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
