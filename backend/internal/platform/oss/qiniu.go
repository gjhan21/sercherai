package oss

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// QiniuConfig represents the configuration for Qiniu OSS.
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string
	UseHTTPS  bool
}

// UploadToQiniu uploads a payload to Qiniu and returns the public URL.
func UploadToQiniu(cfg QiniuConfig, objectKey string, payload []byte, mimeType string) (string, error) {
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" ||
		strings.TrimSpace(cfg.Bucket) == "" || strings.TrimSpace(cfg.Domain) == "" {
		return "", errors.New("qiniu config incomplete, require access_key/secret_key/bucket/domain")
	}
	if len(payload) == 0 {
		return "", errors.New("empty upload payload")
	}

	token, err := BuildQiniuUploadToken(cfg.AccessKey, cfg.SecretKey, cfg.Bucket, objectKey)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("token", token); err != nil {
		return "", err
	}
	if err := writer.WriteField("key", objectKey); err != nil {
		return "", err
	}
	part, err := writer.CreateFormFile("file", filepath.Base(objectKey))
	if err != nil {
		return "", err
	}
	if _, err := part.Write(payload); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	uploadHost := BuildQiniuUploadHost(cfg.Region, cfg.UseHTTPS)
	req, err := http.NewRequest(http.MethodPost, uploadHost, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(mimeType) != "" {
		req.Header.Set("X-Qiniu-MimeType", mimeType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBytes))
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("qiniu upload failed: %s", msg)
	}

	if len(respBytes) > 0 {
		var result map[string]interface{}
		if json.Unmarshal(respBytes, &result) == nil {
			if errText, ok := result["error"].(string); ok && strings.TrimSpace(errText) != "" {
				return "", fmt.Errorf("qiniu upload failed: %s", strings.TrimSpace(errText))
			}
			if key, ok := result["key"].(string); ok && strings.TrimSpace(key) != "" {
				objectKey = strings.TrimSpace(key)
			}
		}
	}

	return BuildQiniuPublicURL(cfg, objectKey), nil
}

// BuildQiniuUploadHost returns the upload URL for the specified region.
func BuildQiniuUploadHost(region string, useHTTPS bool) string {
	regionHost := map[string]string{
		"z0":  "up-z0.qiniup.com",
		"z1":  "up-z1.qiniup.com",
		"z2":  "up-z2.qiniup.com",
		"na0": "up-na0.qiniup.com",
		"as0": "up-as0.qiniup.com",
	}
	host, ok := regionHost[strings.ToLower(strings.TrimSpace(region))]
	if !ok || host == "" {
		host = "up-z0.qiniup.com"
	}
	scheme := "https://"
	if !useHTTPS {
		scheme = "http://"
	}
	return scheme + host
}

// BuildQiniuPublicURL constructs the public URL for an object.
func BuildQiniuPublicURL(cfg QiniuConfig, objectKey string) string {
	domain := strings.TrimSpace(cfg.Domain)
	if domain == "" {
		return objectKey
	}
	if strings.HasPrefix(strings.ToLower(domain), "http://") || strings.HasPrefix(strings.ToLower(domain), "https://") {
		return strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(objectKey, "/")
	}
	scheme := "https://"
	if !cfg.UseHTTPS {
		scheme = "http://"
	}
	return scheme + strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

// BuildQiniuUploadToken generates a Qiniu upToken.
func BuildQiniuUploadToken(accessKey, secretKey, bucket, objectKey string) (string, error) {
	ak := strings.TrimSpace(accessKey)
	sk := strings.TrimSpace(secretKey)
	bk := strings.TrimSpace(bucket)
	key := strings.TrimSpace(objectKey)
	if ak == "" || sk == "" || bk == "" || key == "" {
		return "", errors.New("qiniu credential or object key missing")
	}

	policy := map[string]interface{}{
		"scope":    bk + ":" + key,
		"deadline": time.Now().Add(1 * time.Hour).Unix(),
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	encodedPolicy := base64.URLEncoding.EncodeToString(policyJSON)
	mac := hmac.New(sha1.New, []byte(sk))
	if _, err := mac.Write([]byte(encodedPolicy)); err != nil {
		return "", err
	}
	sign := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return ak + ":" + sign + ":" + encodedPolicy, nil
}
