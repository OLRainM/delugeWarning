package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTTSEngine 调用自建 TTS HTTP 服务（OpenAI 风格 /v1/audio/speech 接口）合成语音，
// 合成后通过 Storage 上传并返回 ObjectKey。
type HTTPTTSEngine struct {
	storage    Storage
	endpoint   string        // 如 http://115.190.125.177:3000/v1/audio/speech
	voice      string        // 如 zh-CN-YunxiNeural
	speed      float64       // 语速，1.0 为正常，0.9 稍慢
	httpClient *http.Client
}

// ttsSpeechReq 对应服务端 /v1/audio/speech 的请求体。
type ttsSpeechReq struct {
	Voice string  `json:"voice"`
	Input string  `json:"input"`
	Speed float64 `json:"speed"`
}

// NewHTTPTTS 构造 HTTPTTSEngine。
// endpoint 为空时使用你自建服务的默认地址；voice 为空时用男声 zh-CN-YunxiNeural；speed 为 0 时用 0.9。
func NewHTTPTTS(storage Storage, endpoint, voice string, speed float64) *HTTPTTSEngine {
	if endpoint == "" {
		endpoint = "http://115.190.125.177:3000/v1/audio/speech"
	}
	if voice == "" {
		voice = "zh-CN-YunxiNeural"
	}
	if speed <= 0 {
		speed = 0.9
	}
	return &HTTPTTSEngine{
		storage:    storage,
		endpoint:   endpoint,
		voice:      voice,
		speed:      speed,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Synthesize 调用 TTS 服务合成语音，上传对象存储后返回 ObjectKey。
func (e *HTTPTTSEngine) Synthesize(ctx context.Context, text string) (string, error) {
	payload := ttsSpeechReq{Voice: e.voice, Input: text, Speed: e.speed}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("请求体编码失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("TTS 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("TTS 服务返回错误，状态码: %d, 信息: %s", resp.StatusCode, string(errBody))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取音频数据失败: %w", err)
	}
	if len(audioData) == 0 {
		return "", fmt.Errorf("TTS 服务返回空音频数据")
	}

	// 上传到对象存储，ObjectKey 用时间戳保证唯一
	objectKey := fmt.Sprintf("audio/tts-%d.mp3", time.Now().UnixNano())
	return e.storage.Upload(ctx, objectKey, audioData, "audio/mpeg")
}
