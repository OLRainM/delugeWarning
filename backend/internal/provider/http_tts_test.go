package provider

import (
	"context"
	"strings"
	"testing"
)

func TestHTTPTTS_Synthesize(t *testing.T) {
	tts := NewHTTPTTS(MockStorage{}, "http://115.190.125.177:3000/v1/audio/speech", "zh-CN-YunxiNeural", 0.9)
	key, err := tts.Synthesize(context.Background(), "橙色预警，请注意防范洪水")
	if err != nil {
		t.Fatalf("合成失败: %v", err)
	}
	if !strings.HasPrefix(key, "audio/tts-") || !strings.HasSuffix(key, ".mp3") {
		t.Fatalf("ObjectKey 格式不符: %s", key)
	}
	t.Logf("✅ 合成成功，ObjectKey=%s", key)
}
