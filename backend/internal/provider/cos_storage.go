package provider

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// normalizeKey 规整 ObjectKey：兼容历史数据里误存为完整 URL 的情况，
// 提取出相对路径（如 audio/tts-xxx.mp3），并去掉开头的斜杠。
func normalizeKey(key string) string {
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		if u, err := url.Parse(key); err == nil {
			key = u.Path
		}
	}
	return strings.TrimPrefix(key, "/")
}

// COSStorage 腾讯云 COS 私有桶实现。
// 数据库只存 ObjectKey（相对路径），下发时通过 GetDownloadURL 签名生成临时链接。
type COSStorage struct {
	client    *cos.Client
	secretID  string
	secretKey string
}

// NewCOSStorage 用自定义域名（HTTPS）初始化 COS 客户端。
func NewCOSStorage(secretID, secretKey, baseURL string) (*COSStorage, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("COS baseURL 解析失败: %w", err)
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return &COSStorage{client: client, secretID: secretID, secretKey: secretKey}, nil
}

// Upload 服务端直传（TTS 合成音频等），返回 ObjectKey。
func (s *COSStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	key = normalizeKey(key)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	_, err := s.client.Object.Put(ctx, key, bytes.NewReader(data), opt)
	if err != nil {
		return "", err
	}
	return key, nil
}

// PresignPut 为前端直传生成预签名 PUT 地址，返回 (uploadURL, objectKey)。
func (s *COSStorage) PresignPut(ctx context.Context, key string, ttl time.Duration) (string, string, error) {
	key = normalizeKey(key)
	presignedURL, err := s.client.Object.GetPresignedURL(
		ctx, http.MethodPut, key, s.secretID, s.secretKey, ttl, nil,
	)
	if err != nil {
		return "", "", err
	}
	return presignedURL.String(), key, nil
}

// GetDownloadURL 对私有桶 ObjectKey 签发临时 GET 链接，有效期由调用方指定（通常 15 分钟）。
func (s *COSStorage) GetDownloadURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	objectKey = normalizeKey(objectKey)
	presignedURL, err := s.client.Object.GetPresignedURL(
		ctx, http.MethodGet, objectKey, s.secretID, s.secretKey, ttl, nil,
	)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
