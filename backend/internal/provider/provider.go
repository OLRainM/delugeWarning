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

// Storage 对象存储抽象（腾讯云 COS，私有桶）。
type Storage interface {
	// Upload 上传文件，返回 ObjectKey（相对路径，不含域名）。服务端 TTS 合成后调用。
	Upload(ctx context.Context, key string, data []byte, contentType string) (objectKey string, err error)
	// PresignPut 返回前端直传用的预签名 PUT 地址及 ObjectKey。
	PresignPut(ctx context.Context, key string, ttl time.Duration) (uploadURL, objectKey string, err error)
	// GetDownloadURL 对私有桶的 ObjectKey 签发临时可访问的 GET 链接。
	GetDownloadURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
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

func (MockStorage) Upload(_ context.Context, key string, _ []byte, _ string) (string, error) {
	return key, nil
}

func (MockStorage) PresignPut(_ context.Context, key string, _ time.Duration) (string, string, error) {
	return "https://mock.local/upload/" + key, key, nil
}

func (MockStorage) GetDownloadURL(_ context.Context, objectKey string, _ time.Duration) (string, error) {
	return "https://mock.local/file/" + objectKey, nil
}

type MockPusher struct{}

func (MockPusher) Push(_ context.Context, openid string, msg TemplateMsg) error {
	fmt.Printf("[mock-push] -> openid=%s tmpl=%s data=%v\n", openid, msg.TemplateID, msg.Data)
	return nil
}
