// verify_cos/main.go —— COS 存储链路端到端验证
// 用法：go run ./verify_cos -sid AKID... -skey Sm2... -bucket leduge-1309942422 -base https://cos.olraingin.com
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"delugewarning/internal/provider"
)

func main() {
	sid := flag.String("sid", os.Getenv("COS_SECRET_ID"), "SecretId")
	skey := flag.String("skey", os.Getenv("COS_SECRET_KEY"), "SecretKey")
	base := flag.String("base", "https://cos.olraingin.com", "桶自定义域名")
	flag.Parse()

	if *sid == "" || *skey == "" {
		fmt.Println("❌ 缺少 SecretId/SecretKey，请用 -sid/-skey 传入或设置环境变量")
		os.Exit(1)
	}

	s, err := provider.NewCOSStorage(*sid, *skey, *base)
	must("初始化 COSStorage", err)
	fmt.Printf("✅ COSStorage 初始化成功，域名=%s\n\n", *base)

	ctx := context.Background()
	key := fmt.Sprintf("verify/test-%d.txt", time.Now().UnixNano())
	content := []byte("deluge-warning COS链路验证 " + time.Now().Format(time.RFC3339))

	// ① Upload：服务端直传
	fmt.Printf("【1/3】Upload → %s\n", key)
	objectKey, err := s.Upload(ctx, key, content, "text/plain")
	must("Upload", err)
	fmt.Printf("✅ 上传成功，ObjectKey=%s\n\n", objectKey)

	// ② GetDownloadURL：签名下载
	fmt.Println("【2/3】GetDownloadURL（有效期15分钟）")
	dlURL, err := s.GetDownloadURL(ctx, objectKey, 15*time.Minute)
	must("GetDownloadURL", err)
	fmt.Printf("✅ 预签名 GET URL 生成成功\n   %s\n\n", dlURL)

	// 用生成的 URL 实际下载并比对内容
	resp, err := http.Get(dlURL) //nolint:noctx
	must("HTTP GET 下载", err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 下载失败，HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}
	got, _ := io.ReadAll(resp.Body)
	if string(got) != string(content) {
		fmt.Printf("❌ 内容不一致\n  期望: %s\n  实际: %s\n", content, got)
		os.Exit(1)
	}
	fmt.Printf("✅ 内容验证通过：\"%s\"\n\n", got)

	// ③ PresignPut：前端直传预签名
	putKey := fmt.Sprintf("verify/presign-%d.txt", time.Now().UnixNano())
	fmt.Printf("【3/3】PresignPut → %s\n", putKey)
	putURL, retKey, err := s.PresignPut(ctx, putKey, 10*time.Minute)
	must("PresignPut", err)
	fmt.Printf("✅ 预签名 PUT URL 生成成功，ObjectKey=%s\n   %s\n\n", retKey, putURL)

	// 用 PUT URL 实际上传一个小文件验证可用性
	req, _ := http.NewRequest(http.MethodPut, putURL, nil)
	putBody := []byte("presign-put-test")
	req.Body = io.NopCloser(bytes.NewReader(putBody))
	req.ContentLength = int64(len(putBody))
	putResp, err := http.DefaultClient.Do(req)
	must("HTTP PUT 直传", err)
	putResp.Body.Close()
	if putResp.StatusCode != 200 {
		fmt.Printf("❌ 直传失败，HTTP %d\n", putResp.StatusCode)
		os.Exit(1)
	}
	fmt.Printf("✅ 前端直传验证通过，HTTP %d\n\n", putResp.StatusCode)

	fmt.Println("════════════════════════════════")
	fmt.Println("✅ COS 存储链路全部验证通过")
}

func must(label string, err error) {
	if err != nil {
		fmt.Printf("❌ %s 失败: %v\n", label, err)
		os.Exit(1)
	}
}
