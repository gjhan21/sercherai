package handler

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const paymentEnabledConfigKey = "payment.enabled"
const paymentDefaultChannelConfigKey = "payment.default_channel"
const paymentSigningSecretConfigKey = "payment.signing_secret"
const paymentChannelYolkPayEnabledConfigKey = "payment.channel.yolkpay.enabled"
const paymentChannelYolkPayPIDConfigKey = "payment.channel.yolkpay.pid"
const paymentChannelYolkPayKeyConfigKey = "payment.channel.yolkpay.key"
const paymentChannelYolkPayGatewayConfigKey = "payment.channel.yolkpay.gateway"
const paymentChannelYolkPayMAPIPathConfigKey = "payment.channel.yolkpay.mapi_path"
const paymentChannelYolkPayNotifyURLConfigKey = "payment.channel.yolkpay.notify_url"
const paymentChannelYolkPayReturnURLConfigKey = "payment.channel.yolkpay.return_url"
const paymentChannelYolkPayPayTypeConfigKey = "payment.channel.yolkpay.pay_type"
const paymentChannelYolkPayDeviceConfigKey = "payment.channel.yolkpay.device"

func buildYolkPaySign(params map[string]string, merchantKey string) string {
	filtered := make([]string, 0, len(params))
	values := make(map[string]string, len(params))
	for rawKey, rawValue := range params {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}
		lowerKey := strings.ToLower(key)
		if lowerKey == "sign" || lowerKey == "sign_type" {
			continue
		}
		value := strings.TrimSpace(rawValue)
		if value == "" {
			continue
		}
		filtered = append(filtered, key)
		values[key] = value
	}
	sort.Strings(filtered)
	parts := make([]string, 0, len(filtered))
	for _, key := range filtered {
		parts = append(parts, key+"="+values[key])
	}
	signSource := strings.Join(parts, "&") + strings.TrimSpace(merchantKey)
	hash := md5.Sum([]byte(signSource))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

func verifyYolkPaySign(params map[string]string, sign string, merchantKey string) bool {
	expected := buildYolkPaySign(params, merchantKey)
	return strings.EqualFold(strings.TrimSpace(sign), strings.TrimSpace(expected))
}

func buildYolkPayGatewayURL(gateway string, mapiPath string) string {
	normalizedGateway := strings.TrimSpace(gateway)
	if normalizedGateway == "" {
		normalizedGateway = "https://www.yolkpay.net"
	}
	lowerGateway := strings.ToLower(normalizedGateway)
	if !strings.HasPrefix(lowerGateway, "http://") && !strings.HasPrefix(lowerGateway, "https://") {
		normalizedGateway = "https://" + normalizedGateway
	}
	normalizedGateway = strings.TrimRight(normalizedGateway, "/")

	normalizedPath := strings.TrimSpace(mapiPath)
	if normalizedPath == "" {
		normalizedPath = "/mapi.php"
	}
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}
	return normalizedGateway + normalizedPath
}

func parseYolkPayCode(raw interface{}) int {
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func stringifyYolkPayValue(raw interface{}) string {
	if raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		text := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if strings.EqualFold(text, "<nil>") {
			return ""
		}
		return text
	}
}
