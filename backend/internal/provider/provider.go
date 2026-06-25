package provider

import (
	"context"
	"fmt"
	"time"
)

// TTSEngine 文本转语音：合成后返回可访问的音频 URL。
type TTSEngine interface {
	Synthesize(ctx context.Context, text string) (audioURL string, err error)
}

// Storage 对象存储抽象（默认腾讯云 COS，可换 S3/OSS/本地）。
type Storage interface {
	// PresignPut 返回前端直传用的预签名地址，以及最终访问 URL。
	PresignPut(ctx context.Context, key string, ttl time.Duration) (uploadURL, accessURL string, err error)
}

// TemplateMsg 微信订阅消息内容。
type TemplateMsg struct {
	TemplateID string
	Page       string
	Data       map[string]string
}

// Pusher 消息推送抽象（默认微信订阅消息）。
type Pusher interface {
	Push(ctx context.Context, openid string, msg TemplateMsg) error
}

// ---------- mock 实现：本地联调用，不依赖外部云服务 ----------

type MockTTS struct{}

func (MockTTS) Synthesize(_ context.Context, text string) (string, error) {
	// 返回占位 URL，实际项目中替换为腾讯云 TTS 合成结果
	return fmt.Sprintf("https://mock.local/tts/%d.mp3", time.Now().UnixNano()), nil
}

type MockStorage struct{}

func (MockStorage) PresignPut(_ context.Context, key string, _ time.Duration) (string, string, error) {
	return "https://mock.local/upload/" + key, "https://mock.local/file/" + key, nil
}

type MockPusher struct{}

func (MockPusher) Push(_ context.Context, openid string, msg TemplateMsg) error {
	fmt.Printf("[mock-push] -> openid=%s tmpl=%s data=%v\n", openid, msg.TemplateID, msg.Data)
	return nil
}
