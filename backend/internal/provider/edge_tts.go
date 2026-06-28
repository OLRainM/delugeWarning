package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// EdgeTTSEngine 使用 edge-tts（免费）合成语音，合成后通过 Storage 上传并返回 ObjectKey。
// 依赖：已安装 edge-tts（pip install edge-tts）且 python/py 可执行。
type EdgeTTSEngine struct {
	storage    Storage
	voice      string // 默认 zh-CN-XiaoxiaoNeural
	pythonBin  string // python 可执行文件名
}

// NewEdgeTTS 构造 EdgeTTSEngine。
// storage 用于把合成的 mp3 上传对象存储，voice 留空时用默认女声。
func NewEdgeTTS(storage Storage, voice, pythonBin string) *EdgeTTSEngine {
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}
	if pythonBin == "" {
		pythonBin = "py" // Windows 用 py，Linux/Mac 用 python3
	}
	return &EdgeTTSEngine{storage: storage, voice: voice, pythonBin: pythonBin}
}

// Synthesize 合成文本为语音，返回存入对象存储的 ObjectKey。
func (e *EdgeTTSEngine) Synthesize(ctx context.Context, text string) (string, error) {
	// 写入临时 mp3 文件
	tmp, err := os.CreateTemp("", "tts-*.mp3")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	// 调用 edge-tts 合成
	cmd := exec.CommandContext(ctx, e.pythonBin,
		"-m", "edge_tts",
		"--voice", e.voice,
		"--text", text,
		"--write-media", tmpPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("edge-tts 合成失败: %w, output: %s", err, out)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("读取合成文件失败: %w", err)
	}

	// 上传到对象存储，ObjectKey 用时间戳保证唯一
	objectKey := fmt.Sprintf("audio/tts-%d%s", time.Now().UnixNano(), filepath.Ext(tmpPath))
	return e.storage.Upload(ctx, objectKey, data, "audio/mpeg")
}
