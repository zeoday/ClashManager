package service

import (
	"clash-manager/internal/model"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ParseSubscription parses a subscription URL and returns a list of nodes
func ParseSubscription(subURL string) ([]model.Node, error) {
	// Fetch subscription content
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", subURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// Set common User-Agent to avoid blocking
	req.Header.Set("User-Agent", "clash-verge/v1.3.8")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// Try to detect content type and parse accordingly
	nodes, err := parseContent(string(content))
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// parseContent tries to parse subscription content in various formats
func parseContent(content string) ([]model.Node, error) {
	content = strings.TrimSpace(content)

	// Try YAML format first (Clash config)
	if strings.HasPrefix(content, "proxies:") || strings.HasPrefix(content, "mixed-port:") {
		return parseYAMLConfig(content)
	}

	// Try Base64 encoded links
	return parseBase64Links(content)
}

// parseYAMLConfig parses a YAML Clash configuration
func parseYAMLConfig(content string) ([]model.Node, error) {
	var config struct {
		Proxies []map[string]interface{} `yaml:"proxies"`
	}

	decoder := yaml.NewDecoder(strings.NewReader(content))
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("parse YAML failed: %w", err)
	}

	var nodes []model.Node
	for _, proxy := range config.Proxies {
		node, err := proxyToNode(proxy)
		if err != nil {
			// Skip invalid nodes but continue parsing
			continue
		}
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

// proxyToNode converts a YAML proxy map to a Node model
func proxyToNode(proxy map[string]interface{}) (*model.Node, error) {
	nodeType, _ := proxy["type"].(string)
	if nodeType == "" {
		return nil, fmt.Errorf("missing proxy type")
	}

	name, _ := proxy["name"].(string)
	if name == "" {
		name = "Unnamed Node"
	}

	node := &model.Node{
		Name:     name,
		Type:     nodeType,
		Server:   getStringVal(proxy, "server"),
		UDP:      getBoolVal(proxy, "udp"),
		TLS:      getBoolVal(proxy, "tls"),
		SkipCert: getBoolVal(proxy, "skip-cert-verify"),
		Network:  getStringVal(proxy, "network"),
		Path:     getStringVal(proxy, "path"),
		Host:     getStringVal(proxy, "host"),
		ALPN:     getStringVal(proxy, "alpn"),
	}

	// Port
	if port, ok := proxy["port"].(int); ok {
		node.Port = port
	} else if portStr, ok := proxy["port"].(string); ok {
		fmt.Sscanf(portStr, "%d", &node.Port)
	}

	// Build extra config for complex fields
	extra := make(map[string]interface{})

	// Type-specific fields
	switch nodeType {
	case "ss":
		node.Cipher = getStringVal(proxy, "cipher")
		node.Password = getStringVal(proxy, "password")
	case "vmess":
		node.UUID = getStringVal(proxy, "uuid")
		node.Cipher = getStringVal(proxy, "cipher")
		if node.Cipher == "" {
			node.Cipher = "auto"
		}
		if aid := getIntVal(proxy, "alterId"); aid > 0 {
			node.Flow = fmt.Sprintf("%d", aid) // Store alterId in Flow temporarily
		}
	case "vless":
		node.UUID = getStringVal(proxy, "uuid")
		if flow := getStringVal(proxy, "flow"); flow != "" {
			node.Flow = flow
		}
		// Reality 配置
		if realityOpts, ok := proxy["reality-opts"].(map[string]interface{}); ok {
			node.RealityPublicKey = getStringVal(realityOpts, "public-key")
			node.RealityShortID = getStringVal(realityOpts, "short-id")
		}
	case "trojan":
		node.Password = getStringVal(proxy, "password")
	case "hysteria2", "hy2":
		node.Type = "hysteria2"
		node.Password = getStringVal(proxy, "password")
		// 带宽配置
		if up := getStringVal(proxy, "up"); up != "" {
			node.UpMbps = parseBandwidth(up)
		}
		if down := getStringVal(proxy, "down"); down != "" {
			node.DownMbps = parseBandwidth(down)
		}
		// 也支持 up-mbps/down-mbps 格式
		if upMbps := getIntVal(proxy, "up-mbps"); upMbps > 0 {
			node.UpMbps = upMbps
		}
		if downMbps := getIntVal(proxy, "down-mbps"); downMbps > 0 {
			node.DownMbps = downMbps
		}
	case "hysteria":
		node.Password = getStringVal(proxy, "auth-str")
		if node.Password == "" {
			node.Password = getStringVal(proxy, "auth")
		}
		if up := getStringVal(proxy, "up"); up != "" {
			node.UpMbps = parseBandwidth(up)
		}
		if down := getStringVal(proxy, "down"); down != "" {
			node.DownMbps = parseBandwidth(down)
		}
	case "socks5":
		node.Username = getStringVal(proxy, "username")
		node.Password = getStringVal(proxy, "password")
	case "http":
		node.Username = getStringVal(proxy, "username")
		node.Password = getStringVal(proxy, "password")
	case "wireguard":
		node.PublicKey = getStringVal(proxy, "public-key")
		node.PrivateKey = getStringVal(proxy, "private-key")
		if mtu := getIntVal(proxy, "mtu"); mtu > 0 {
			node.MTU = mtu
		}
		if ip := getStringVal(proxy, "ip"); ip != "" {
			extra["ip"] = ip
		}
		if ipv6 := getStringVal(proxy, "ipv6"); ipv6 != "" {
			extra["ipv6"] = ipv6
		}
	case "tuic":
		node.UUID = getStringVal(proxy, "uuid")
		node.Password = getStringVal(proxy, "password")
		node.CongestionControl = getStringVal(proxy, "congestion-control")
		if node.CongestionControl == "" {
			node.CongestionControl = "cubic"
		}
	}

	// WebSocket 配置解析
	if wsOpts, ok := proxy["ws-opts"].(map[string]interface{}); ok {
		if node.Network == "" {
			node.Network = "ws"
		}
		node.Path = getStringVal(wsOpts, "path")
		node.MaxEarlyData = getIntVal(wsOpts, "max-early-data")
		node.EarlyDataHeader = getStringVal(wsOpts, "early-data-header-name")
		if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
			if host := getStringVal(headers, "Host"); host != "" {
				node.Host = host
			}
		}
	}

	// gRPC 配置解析
	if grpcOpts, ok := proxy["grpc-opts"].(map[string]interface{}); ok {
		if node.Network == "" {
			node.Network = "grpc"
		}
		node.ServiceName = getStringVal(grpcOpts, "grpc-service-name")
	}

	// HTTP/2 配置解析
	if h2Opts, ok := proxy["h2-opts"].(map[string]interface{}); ok {
		if node.Network == "" {
			node.Network = "h2"
		}
		node.Path = getStringVal(h2Opts, "path")
		if hosts, ok := h2Opts["host"].([]interface{}); ok && len(hosts) > 0 {
			if host, ok := hosts[0].(string); ok {
				node.Host = host
			}
		}
	}

	// TLS 配置
	if tlsOpts, ok := proxy["tls-opts"].(map[string]interface{}); ok {
		if sni := getStringVal(tlsOpts, "sni"); sni != "" && node.Host == "" {
			node.Host = sni
		}
	}

	// SNI 也可以直接在 proxy 级别
	if sni := getStringVal(proxy, "sni"); sni != "" && node.Host == "" {
		node.Host = sni
	}

	// Client Fingerprint (UTLS)
	if clientFingerprint := getStringVal(proxy, "client-fingerprint"); clientFingerprint != "" {
		node.ClientFingerprint = clientFingerprint
	}

	// Service Name (gRPC 的另一种写法)
	if serviceName := getStringVal(proxy, "service-name"); serviceName != "" && node.ServiceName == "" {
		node.ServiceName = serviceName
	}

	// Multiplex 支持
	if getBoolVal(proxy, "smux") || getBoolVal(proxy, "multiplex") {
		node.Multiplex = true
	}

	if len(extra) > 0 {
		if b, err := yaml.Marshal(extra); err == nil {
			node.ExtraConfig = string(b)
		}
	}

	return node, nil
}

// parseBandwidth parses bandwidth string like "100 Mbps" or "100" to int
func parseBandwidth(s string) int {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Remove common suffixes
	s = strings.TrimSuffix(s, "mbps")
	s = strings.TrimSuffix(s, "mb")
	s = strings.TrimSuffix(s, "mb/s")
	s = strings.TrimSuffix(s, "m")
	s = strings.TrimSpace(s)

	var mbps int
	fmt.Sscanf(s, "%d", &mbps)
	return mbps
}

// parseBase64Links parses Base64 encoded subscription links
func parseBase64Links(content string) ([]model.Node, error) {
	var nodes []model.Node

	// Try to decode Base64
	decoded, err := tryBase64Decode(content)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	// Split by newlines and parse each link
	links := strings.Split(strings.TrimSpace(decoded), "\n")

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}

		node, err := ParseLink(link)
		if err != nil {
			// Try parsing as VLESS or other protocols
			if node, err = parseAlternativeLink(link); err != nil {
				continue
			}
		}
		nodes = append(nodes, *node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid nodes found in subscription")
	}

	return nodes, nil
}

// tryBase64Decode tries multiple Base64 decoding methods
func tryBase64Decode(content string) (string, error) {
	// Method 1: Standard Base64
	if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
		return string(decoded), nil
	}

	// Method 2: URL-safe Base64
	if decoded, err := base64.RawURLEncoding.DecodeString(content); err == nil {
		return string(decoded), nil
	}

	// Method 3: Raw Standard Base64
	if decoded, err := base64.RawStdEncoding.DecodeString(content); err == nil {
		return string(decoded), nil
	}

	// Method 4: Try with padding
	content = strings.TrimRight(content, "=")
	padded := content + strings.Repeat("=", (4-len(content)%4)%4)
	if decoded, err := base64.StdEncoding.DecodeString(padded); err == nil {
		return string(decoded), nil
	}

	return "", fmt.Errorf("all decode methods failed")
}

// parseAlternativeLink parses VLESS and other alternative link formats
func parseAlternativeLink(link string) (*model.Node, error) {
	if strings.HasPrefix(link, "vless://") {
		return parseVLESS(link)
	}
	if strings.HasPrefix(link, "hysteria://") || strings.HasPrefix(link, "hysteria2://") {
		return parseHysteria(link)
	}
	if strings.HasPrefix(link, "socks://") || strings.HasPrefix(link, "socks5://") {
		return parseSocks(link)
	}
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return parseHTTPProxy(link)
	}
	return nil, fmt.Errorf("unsupported link format")
}

// parseVLESS parses vless:// links
func parseVLESS(link string) (*model.Node, error) {
	// vless://uuid@host:port?params#name
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	node := &model.Node{
		Type:   "vless",
		Name:   u.Fragment,
		UUID:   u.User.Username(),
		Server: u.Hostname(),
	}

	if node.Name == "" {
		node.Name = "VLESS Node"
	}

	node.Port, _ = strconv.Atoi(u.Port())

	// Parse query parameters
	q := u.Query()
	node.Network = q.Get("type")
	security := q.Get("security")
	node.TLS = security == "tls" || security == "reality"
	node.Host = q.Get("sni")
	node.Path = q.Get("path")

	// Flow (XTLS)
	if flow := q.Get("flow"); flow != "" {
		node.Flow = flow
	}

	// Reality 配置
	if security == "reality" {
		node.RealityPublicKey = q.Get("pbk")
		if node.RealityPublicKey == "" {
			node.RealityPublicKey = q.Get("publicKey")
		}
		node.RealityShortID = q.Get("sid")
		if node.RealityShortID == "" {
			node.RealityShortID = q.Get("shortId")
		}
		if node.Host == "" {
			node.Host = q.Get("sni")
		}
	}

	// Client Fingerprint
	if fp := q.Get("fp"); fp != "" {
		node.ClientFingerprint = fp
	}

	// WebSocket 配置
	if node.Network == "ws" {
		node.Path = q.Get("path")
		if host := q.Get("host"); host != "" {
			node.Host = host
		}
		if ed := q.Get("ed"); ed != "" {
			node.MaxEarlyData, _ = strconv.Atoi(ed)
		}
	}

	// gRPC 配置
	if node.Network == "grpc" {
		node.ServiceName = q.Get("serviceName")
		if node.ServiceName == "" {
			node.ServiceName = q.Get("service-name")
		}
	}

	return node, nil
}

// parseHysteria parses hysteria:// and hysteria2:// links
func parseHysteria(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	node := &model.Node{
		Name:   u.Fragment,
		Server: u.Hostname(),
	}

	if node.Name == "" {
		node.Name = "Hysteria2 Node"
	}

	node.Port, _ = strconv.Atoi(u.Port())

	q := u.Query()

	// 区分 Hysteria v1 和 v2
	if strings.HasPrefix(link, "hysteria2://") || strings.HasPrefix(link, "hy2://") {
		node.Type = "hysteria2"
		node.Password = u.User.Username()
		if node.Password == "" {
			node.Password = q.Get("auth")
		}
	} else {
		node.Type = "hysteria"
		node.Password = q.Get("auth")
		if node.Password == "" {
			node.Password = q.Get("auth-str")
		}
	}

	// SNI
	node.Host = q.Get("sni")
	if node.Host == "" {
		node.Host = q.Get("peer")
	}

	// 跳过证书验证
	if q.Get("insecure") == "1" || q.Get("insecure") == "true" {
		node.SkipCert = true
	}

	// 带宽配置
	if up := q.Get("up"); up != "" {
		node.UpMbps = parseBandwidth(up)
	}
	if down := q.Get("down"); down != "" {
		node.DownMbps = parseBandwidth(down)
	}
	// 也支持 upmbps/downmbps 格式
	if upmbps := q.Get("upmbps"); upmbps != "" {
		node.UpMbps = parseBandwidth(upmbps)
	}
	if downmbps := q.Get("downmbps"); downmbps != "" {
		node.DownMbps = parseBandwidth(downmbps)
	}

	// Obfs (Hysteria v1)
	if obfs := q.Get("obfs"); obfs != "" {
		extra := make(map[string]interface{})
		extra["obfs"] = obfs
		if b, err := yaml.Marshal(extra); err == nil {
			node.ExtraConfig = string(b)
		}
	}

	return node, nil
}

// parseSocks parses socks:// and socks5:// links
func parseSocks(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	node := &model.Node{
		Type:   "socks5",
		Name:   u.Fragment,
		Server: u.Hostname(),
	}

	if node.Name == "" {
		node.Name = "SOCKS5 Node"
	}

	node.Port, _ = strconv.Atoi(u.Port())

	if u.User != nil {
		node.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			node.Password = password
		}
	}

	return node, nil
}

// parseHTTPProxy parses http:// and https:// proxy links
func parseHTTPProxy(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	node := &model.Node{
		Type:   "http",
		Name:   u.Fragment,
		Server: u.Hostname(),
		TLS:    u.Scheme == "https",
	}

	if node.Name == "" {
		node.Name = "HTTP Node"
	}

	node.Port, _ = strconv.Atoi(u.Port())

	if u.User != nil {
		node.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			node.Password = password
		}
	}

	return node, nil
}

// Helper functions
func getStringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBoolVal(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getIntVal(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case string:
			var i int
			fmt.Sscanf(val, "%d", &i)
			return i
		}
	}
	return 0
}

// MergeNodes merges subscription nodes with existing nodes based on sync mode
func MergeNodes(existingNodes []model.Node, newNodes []model.Node, syncMode string, sourceName string) []model.Node {
	var result []model.Node

	switch syncMode {
	case "replace":
		// Replace all: use only new nodes
		result = newNodes

	case "append":
		// Append: keep existing and add new ones
		result = append(result, existingNodes...)
		existingNames := make(map[string]bool)
		for _, n := range existingNodes {
			existingNames[n.Name] = true
		}
		for _, node := range newNodes {
			if !existingNames[node.Name] {
				result = append(result, node)
			}
		}

	case "smart":
		// Smart merge: update existing by name, add new ones
		existingMap := make(map[string]*model.Node)
		for i := range existingNodes {
			existingMap[existingNodes[i].Name] = &existingNodes[i]
		}

		// Start with existing nodes
		result = append(result, existingNodes...)

		for _, newNode := range newNodes {
			if existing, exists := existingMap[newNode.Name]; exists {
				// Update existing node
				newNode.ID = existing.ID
				newNode.CreatedAt = existing.CreatedAt
				// Replace in result
				for i, n := range result {
					if n.Name == newNode.Name {
						result[i] = newNode
						break
					}
				}
			} else {
				// Add new node
				result = append(result, newNode)
			}
		}
	}

	return result
}
