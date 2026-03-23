package service

import (
	"fmt"
	"strings"
)

// ClashToSingBoxConverter handles conversion from Clash config to sing-box format
type ClashToSingBoxConverter struct{}

// NewClashToSingBoxConverter creates a new converter
func NewClashToSingBoxConverter() *ClashToSingBoxConverter {
	return &ClashToSingBoxConverter{}
}

// FieldMapping defines Clash to sing-box field mappings
var fieldMappings = map[string]string{
	// Common fields
	"cipher":           "method",
	"sni":              "server_name",
	"skip-cert-verify": "insecure",
	"servername":       "server_name",

	// VMess fields
	"alterId": "alter_id",
	"alterid": "alter_id",

	// Transport fields
	"grpc-service-name": "service_name",
	"service-name":      "service_name",

	// WebSocket fields
	"max-early-data":         "max_early_data",
	"early-data-header-name": "early_data_header_name",

	// Reality fields
	"public-key": "public_key",
	"short-id":   "short_id",

	// Client fingerprint
	"client-fingerprint": "fingerprint",

	// Authentication
	"auth-str": "auth_str",

	// Hysteria2
	"up-mbps":   "up_mbps",
	"down-mbps": "down_mbps",
}

// ProxyTypeMapping maps Clash proxy types to sing-box types
var proxyTypeMapping = map[string]string{
	"ss":          "shadowsocks",
	"shadowsocks": "shadowsocks",
	"vmess":       "vmess",
	"vless":       "vless",
	"trojan":      "trojan",
	"hysteria2":   "hysteria2",
	"hy2":         "hysteria2",
	"hysteria":    "hysteria",
	"tuic":        "tuic",
	"wireguard":   "wireguard",
	"socks5":      "socks",
	"socks":       "socks",
	"http":        "http",
}

// ConvertProxyType converts Clash proxy type to sing-box type
func (c *ClashToSingBoxConverter) ConvertProxyType(clashType string) string {
	if singboxType, ok := proxyTypeMapping[strings.ToLower(clashType)]; ok {
		return singboxType
	}
	return clashType
}

// ConvertField converts a Clash field name to sing-box field name
func (c *ClashToSingBoxConverter) ConvertField(clashField string) string {
	lowerField := strings.ToLower(clashField)
	if singboxField, ok := fieldMappings[lowerField]; ok {
		return singboxField
	}
	return clashField
}

// ConvertSkipCertVerify converts skip-cert-verify to insecure
func (c *ClashToSingBoxConverter) ConvertSkipCertVerify(skipCertVerify bool) map[string]interface{} {
	return map[string]interface{}{
		"insecure": skipCertVerify,
	}
}

// ConvertNetworkType converts Clash network type
func (c *ClashToSingBoxConverter) ConvertNetworkType(network string) string {
	switch strings.ToLower(network) {
	case "h2":
		return "http"
	default:
		return network
	}
}

// RuleTypeMapping maps Clash rule types to sing-box rule types
var ruleTypeMapping = map[string]string{
	"DOMAIN":         "domain",
	"DOMAIN-SUFFIX":  "domain_suffix",
	"DOMAIN-KEYWORD": "domain_keyword",
	"DOMAIN-REGEX":   "domain_regex",
	"GEOSITE":        "geosite",
	"GEOIP":          "geoip",
	"SRC-GEOIP":      "source_geoip",
	"IP-CIDR":        "ip_cidr",
	"IP-CIDR6":       "ip_cidr",
	"SRC-IP-CIDR":    "source_ip_cidr",
	"SRC-PORT":       "source_port",
	"DST-PORT":       "port",
	"RULE-SET":       "rule_set",
	"PROCESS-NAME":   "process_name",
	"PROCESS-PATH":   "process_path",
	"NETWORK":        "network",
}

// ConvertRuleType converts Clash rule type to sing-box format
func (c *ClashToSingBoxConverter) ConvertRuleType(clashType string) string {
	if singboxType, ok := ruleTypeMapping[strings.ToUpper(clashType)]; ok {
		return singboxType
	}
	return strings.ToLower(clashType)
}

// ConvertRule converts a Clash rule string to sing-box format
// Clash format: TYPE,CONTENT,TARGET
// Returns: rule type, content, target
func (c *ClashToSingBoxConverter) ConvertRule(clashRule string) (ruleType string, content string, target string) {
	parts := strings.Split(clashRule, ",")
	if len(parts) < 2 {
		return "", "", ""
	}

	clashType := strings.TrimSpace(parts[0])
	content = strings.TrimSpace(parts[1])
	if len(parts) >= 3 {
		target = strings.TrimSpace(parts[2])
	}

	ruleType = c.ConvertRuleType(clashType)
	return ruleType, content, target
}

// OutboundTargetMapping maps Clash outbound targets to sing-box
var outboundTargetMapping = map[string]string{
	"DIRECT":      "direct",
	"REJECT":      "block",
	"REJECT-DROP": "block",
	"PASS":        "direct",
}

// ConvertOutboundTarget converts Clash outbound target to sing-box
func (c *ClashToSingBoxConverter) ConvertOutboundTarget(target string) string {
	if singboxTarget, ok := outboundTargetMapping[strings.ToUpper(target)]; ok {
		return singboxTarget
	}
	return target
}

// ConvertALPN converts ALPN from comma-separated string to array
func (c *ClashToSingBoxConverter) ConvertALPN(alpn string) []string {
	if alpn == "" {
		return nil
	}
	parts := strings.Split(alpn, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ConvertBandwidth converts bandwidth string to mbps integer
func (c *ClashToSingBoxConverter) ConvertBandwidth(bandwidth string) int {
	bandwidth = strings.TrimSpace(bandwidth)
	bandwidth = strings.ToLower(bandwidth)

	// Remove common suffixes
	bandwidth = strings.TrimSuffix(bandwidth, "mbps")
	bandwidth = strings.TrimSuffix(bandwidth, "mb")
	bandwidth = strings.TrimSuffix(bandwidth, "mb/s")
	bandwidth = strings.TrimSuffix(bandwidth, "m")
	bandwidth = strings.TrimSpace(bandwidth)

	var mbps int
	if _, err := fmt.Sscanf(bandwidth, "%d", &mbps); err == nil {
		return mbps
	}
	return 0
}

// ConvertGroupType converts Clash proxy group type to sing-box outbound type
func (c *ClashToSingBoxConverter) ConvertGroupType(groupType string) string {
	switch strings.ToLower(groupType) {
	case "select":
		return "selector"
	case "url-test":
		return "urltest"
	case "fallback":
		return "urltest"
	case "load-balance":
		return "urltest"
	default:
		return "selector"
	}
}
