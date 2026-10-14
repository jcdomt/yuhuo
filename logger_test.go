package yuhuo_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jcdomt/yuhuo"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	fn()
	writer.Close()

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestDefaultLoggerDefaultsToInfo(t *testing.T) {
	logger := yuhuo.GetDefaultLogger()

	output := captureStdout(t, func() {
		logger.Debug("debug")
		logger.Info("info")
	})
	if strings.Contains(output, "[DEBUG]") {
		t.Fatalf("默认级别（INFO）不应输出 Debug 日志, got %q", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Fatalf("默认级别（INFO）应输出 Info 日志, got %q", output)
	}
}

func TestDefaultLoggerFiltersLevelsBelowThreshold(t *testing.T) {
	logger := yuhuo.GetDefaultLogger()
	logger.SetLevel(yuhuo.LogLevelWarn)

	output := captureStdout(t, func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error")
	})

	if strings.Contains(output, "[DEBUG]") || strings.Contains(output, "[INFO]") {
		t.Fatalf("WARN 级别不应输出 Debug/Info 日志, got %q", output)
	}
	if !strings.Contains(output, "[WARN]") || !strings.Contains(output, "[ERROR]") {
		t.Fatalf("WARN 级别应输出 Warn/Error 日志, got %q", output)
	}
}

func TestDefaultLoggerInfoLevelSuppressesDebug(t *testing.T) {
	logger := yuhuo.GetDefaultLogger()
	logger.SetLevel(yuhuo.LogLevelInfo)

	output := captureStdout(t, func() {
		logger.Debug("debug")
		logger.Info("info")
	})

	if strings.Contains(output, "[DEBUG]") {
		t.Fatalf("INFO 级别不应输出 Debug 日志, got %q", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Fatalf("INFO 级别应输出 Info 日志, got %q", output)
	}
}

func TestDefaultLoggerIgnoresUnknownLevel(t *testing.T) {
	logger := yuhuo.GetDefaultLogger()
	logger.SetLevel(yuhuo.LogLevelError)
	logger.SetLevel("VERBOSE")

	output := captureStdout(t, func() { logger.Error("error") })
	if !strings.Contains(output, "[ERROR]") {
		t.Fatalf("设置未知级别后应保持原级别, got %q", output)
	}
}

func TestDefaultLoggerFatalPanics(t *testing.T) {
	logger := yuhuo.GetDefaultLogger()
	logger.SetLevel(yuhuo.LogLevelFatal)

	defer func() {
		if recover() == nil {
			t.Fatal("Fatal 应触发 panic")
		}
	}()
	logger.Fatal("fatal")
}
