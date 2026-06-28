package provider

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

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
	presignedURL, err := s.client.Object.GetPresignedURL(
		ctx, http.MethodGet, objectKey, s.secretID, s.secretKey, ttl, nil,
	)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
