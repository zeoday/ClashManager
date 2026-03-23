package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"clash-manager/internal/model"
	"clash-manager/internal/repository"
)

type SingBoxConfigService struct {
	NodeRepo     *repository.NodeRepository
	RuleRepo     *repository.RuleRepository
	GroupRepo    *repository.GroupRepository
	SettingsRepo *repository.SettingsRepository
}

func NewSingBoxConfigService() *SingBoxConfigService {
	return &SingBoxConfigService{
		NodeRepo:     &repository.NodeRepository{},
		RuleRepo:     &repository.RuleRepository{},
		GroupRepo:    &repository.GroupRepository{},
		SettingsRepo: &repository.SettingsRepository{},
	}
}

// GenerateConfig generates sing-box configuration in JSON format
func (s *SingBoxConfigService) GenerateConfig() ([]byte, error) {
	// 1. Fetch Data
	nodes, err := s.NodeRepo.FindAll()
	if err != nil {
		return nil, err
	}
	rules, err := s.RuleRepo.FindAll()
	if err != nil {
		return nil, err
	}
	groups, err := s.GroupRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Build ID Maps
	nodeMap := make(map[uint]string)
	for _, n := range nodes {
		nodeMap[n.ID] = n.Name
	}
	groupMap := make(map[uint]string)
	for _, g := range groups {
		groupMap[g.ID] = g.Name
	}

	// 2. Build Config
	config := model.SingBoxConfig{
		Log: &model.SingBoxLog{
			Level:     "info",
			Timestamp: true,
		},
	}

	// 3. DNS Config
	config.DNS = s.buildDNSConfig()

	// 4. Inbounds (mixed)
	config.Inbounds = []model.SingBoxInbound{
		{
			Type:                     "mixed",
			Tag:                      "mixed-in",
			Listen:                   "0.0.0.0",
			ListenPort:               7893,
			Sniff:                    true,
			SniffOverrideDestination: false,
			DomainStrategy:           "ipv4_only",
			TCPFastOpen:              true,
			UDPFragment:              true,
			SetSystemProxy:           false,
		},
	}

	// 5. Outbounds (nodes + groups)
	outbounds, nodeTags, groupOutMap := s.buildOutbounds(nodes, groups, nodeMap)
	config.Outbounds = outbounds

	// 6. Build Route Rules
	config.Route = s.buildRoute(rules, nodeMap, groupMap, nodeTags, groupOutMap)

	// 7. Marshal to JSON
	return json.MarshalIndent(config, "", "  ")
}

// buildDNSConfig builds DNS configuration for sing-box
func (s *SingBoxConfigService) buildDNSConfig() *model.SingBoxDNS {
	// Try to get DNS config from settings
	dnsVal, err := s.SettingsRepo.Get("dns_config")
	if err == nil && dnsVal != "" {
		var dbDNS model.DNSConfig
		if err := json.Unmarshal([]byte(dnsVal), &dbDNS); err == nil {
			// Convert Clash DNS config to sing-box format
			return s.convertClashDNSToSingBox(&dbDNS)
		}
	}

	// Default DNS config
	return &model.SingBoxDNS{
		Servers: []model.SingBoxDNSServer{
			{
				Tag:     "google",
				Address: "https://8.8.8.8/dns-query",
				Detour:  "proxy",
			},
			{
				Tag:     "local",
				Address: "223.5.5.5",
				Detour:  "direct",
			},
			{
				Tag:     "block",
				Address: "rcode://success",
			},
		},
		Rules: []model.SingBoxDNSRule{
			{
				Outbound: "any",
				Server:   "local",
				Geosite:  []string{"cn"},
			},
		},
		Final:    "google",
		Strategy: "ipv4_only",
	}
}

// convertClashDNSToSingBox converts Clash DNS config to sing-box format
func (s *SingBoxConfigService) convertClashDNSToSingBox(clashDNS *model.DNSConfig) *model.SingBoxDNS {
	dns := &model.SingBoxDNS{
		Servers:  []model.SingBoxDNSServer{},
		Rules:    []model.SingBoxDNSRule{},
		Strategy: "ipv4_only",
	}

	// Add nameservers
	for i, ns := range clashDNS.Nameserver {
		dns.Servers = append(dns.Servers, model.SingBoxDNSServer{
			Tag:     fmt.Sprintf("nameserver-%d", i),
			Address: ns,
			Detour:  "direct",
		})
	}

	// Add fallback servers
	for i, fb := range clashDNS.Fallback {
		dns.Servers = append(dns.Servers, model.SingBoxDNSServer{
			Tag:     fmt.Sprintf("fallback-%d", i),
			Address: fb,
			Detour:  "proxy",
		})
	}

	// Set final
	if len(clashDNS.Fallback) > 0 {
		dns.Final = "fallback-0"
	} else if len(clashDNS.Nameserver) > 0 {
		dns.Final = "nameserver-0"
	}

	return dns
}

// buildOutbounds builds outbounds from nodes and groups
// Returns: outbounds list, node ID -> tag map, group name -> tag map
func (s *SingBoxConfigService) buildOutbounds(nodes []model.Node, groups []model.ProxyGroupModel, nodeMap map[uint]string) ([]model.SingBoxOutbound, map[uint]string, map[string]string) {
	outbounds := []model.SingBoxOutbound{}
	nodeTags := make(map[uint]string)      // node ID -> outbound tag
	groupOutMap := make(map[string]string) // group name -> outbound tag

	// Built-in outbounds
	outbounds = append(outbounds,
		model.SingBoxOutbound{
			Type: "direct",
			Tag:  "direct",
		},
		model.SingBoxOutbound{
			Type: "block",
			Tag:  "block",
		},
		model.SingBoxOutbound{
			Type: "dns",
			Tag:  "dns-out",
		},
	)

	// Convert nodes to outbounds
	for _, n := range nodes {
		outbound := s.nodeToOutbound(&n)
		if outbound != nil {
			outbounds = append(outbounds, *outbound)
			nodeTags[n.ID] = outbound.Tag
		}
	}

	// Collect all node names for default groups
	var allNodeTags []string
	for _, n := range nodes {
		allNodeTags = append(allNodeTags, sanitizeTag(n.Name))
	}

	// Build user-defined groups as selector/urltest outbounds
	for _, g := range groups {
		groupTag := sanitizeTag(g.Name)
		groupOutMap[g.Name] = groupTag

		// Resolve proxy IDs to tags
		var outTags []string
		if g.ProxyIDs != "" {
			var nodeIDs []uint
			if err := json.Unmarshal([]byte(g.ProxyIDs), &nodeIDs); err == nil {
				for _, id := range nodeIDs {
					if tag, ok := nodeTags[id]; ok {
						outTags = append(outTags, tag)
					}
				}
			}
		}

		// Create group outbound based on type
		switch g.Type {
		case "url-test":
			outbounds = append(outbounds, model.SingBoxOutbound{
				Type:      "urltest",
				Tag:       groupTag,
				Outbounds: outTags,
				URL:       s.getTestURL(g.URL),
				Interval:  g.Interval,
				Tolerance: 50,
			})
		case "fallback":
			// fallback in sing-box is also urltest but with different behavior
			outbounds = append(outbounds, model.SingBoxOutbound{
				Type:      "urltest",
				Tag:       groupTag,
				Outbounds: outTags,
				URL:       s.getTestURL(g.URL),
				Interval:  g.Interval,
			})
		case "load-balance":
			// load-balance in sing-box
			outbounds = append(outbounds, model.SingBoxOutbound{
				Type:      "urltest",
				Tag:       groupTag,
				Outbounds: outTags,
				URL:       s.getTestURL(g.URL),
				Interval:  g.Interval,
			})
		default: // select
			outbounds = append(outbounds, model.SingBoxOutbound{
				Type:      "selector",
				Tag:       groupTag,
				Outbounds: outTags,
				Default:   outTags[0],
			})
		}
	}

	// Create default "Auto Select" group (urltest)
	if len(allNodeTags) > 0 {
		autoTag := "Auto_Select"
		groupOutMap["Auto Select"] = autoTag
		outbounds = append(outbounds, model.SingBoxOutbound{
			Type:      "urltest",
			Tag:       autoTag,
			Outbounds: allNodeTags,
			URL:       "http://www.gstatic.com/generate_204",
			Interval:  300,
		})
	}

	// Create default "Proxy" group (selector)
	if len(allNodeTags) > 0 {
		proxyTag := "proxy"
		groupOutMap["Proxy"] = proxyTag
		// Add Auto_Select as first option
		proxyOutTags := append([]string{"Auto_Select"}, allNodeTags...)
		outbounds = append(outbounds, model.SingBoxOutbound{
			Type:      "selector",
			Tag:       proxyTag,
			Outbounds: proxyOutTags,
			Default:   "Auto_Select",
		})
	}

	return outbounds, nodeTags, groupOutMap
}

// getTestURL returns test URL with default fallback
func (s *SingBoxConfigService) getTestURL(url string) string {
	if url == "" {
		return "http://www.gstatic.com/generate_204"
	}
	return url
}

// nodeToOutbound converts a Node to sing-box Outbound
func (s *SingBoxConfigService) nodeToOutbound(node *model.Node) *model.SingBoxOutbound {
	if node == nil {
		return nil
	}

	outbound := &model.SingBoxOutbound{
		Tag:        sanitizeTag(node.Name),
		Server:     node.Server,
		ServerPort: node.Port,
	}

	// TLS configuration
	if node.TLS || node.RealityPublicKey != "" {
		outbound.TLS = &model.SingBoxTLS{
			Enabled:    true,
			ServerName: node.Host,
			Insecure:   node.SkipCert,
		}
		if node.ALPN != "" {
			outbound.TLS.ALPN = strings.Split(node.ALPN, ",")
		}

		// Reality 配置
		if node.RealityPublicKey != "" {
			outbound.TLS.Reality = &model.SingBoxReality{
				Enabled:   true,
				PublicKey: node.RealityPublicKey,
				ShortId:   node.RealityShortID,
			}
		}

		// UTLS 指纹
		if node.ClientFingerprint != "" {
			outbound.TLS.UTLS = &model.SingBoxUTLS{
				Enabled:     true,
				Fingerprint: node.ClientFingerprint,
			}
		}
	}

	// Transport configuration
	if node.Network != "" {
		outbound.Transport = s.buildTransport(node)
	}

	// Multiplex 配置
	if node.Multiplex {
		outbound.Multiplex = &model.SingBoxMultiplex{
			Enabled:  true,
			Protocol: "h2mux",
		}
	}

	// Protocol specific configuration
	switch node.Type {
	case "ss", "shadowsocks":
		outbound.Type = "shadowsocks"
		outbound.Method = node.Cipher
		outbound.Password = node.Password

	case "vmess":
		outbound.Type = "vmess"
		outbound.UUID = node.UUID
		outbound.AlterId = 0
		// 从 Flow 字段获取 alterId（临时存储）
		if node.Flow != "" {
			if aid, err := fmt.Sscanf(node.Flow, "%d", new(int)); err == nil && aid == 1 {
				var aidVal int
				fmt.Sscanf(node.Flow, "%d", &aidVal)
				outbound.AlterId = aidVal
			}
		}
		if node.TLS {
			outbound.Security = "auto"
		}

	case "vless":
		outbound.Type = "vless"
		outbound.UUID = node.UUID
		// VLESS flow (for Reality/XTLS)
		if node.Flow != "" {
			outbound.Flow = node.Flow
		}

	case "trojan":
		outbound.Type = "trojan"
		outbound.Password = node.Password

	case "hysteria2":
		outbound.Type = "hysteria2"
		outbound.Password = node.Password
		// Hysteria2 带宽配置
		if node.UpMbps > 0 || node.DownMbps > 0 {
			outbound.Brutal = &model.SingBoxBrutal{
				Enabled: true,
				Up:      fmt.Sprintf("%d Mbps", node.UpMbps),
				Down:    fmt.Sprintf("%d Mbps", node.DownMbps),
			}
		}

	case "hysteria":
		outbound.Type = "hysteria"
		outbound.AuthString = node.Password
		outbound.UpMbps = node.UpMbps
		outbound.DownMbps = node.DownMbps
		// Hysteria v1 需要 TLS
		if outbound.TLS == nil && node.Host != "" {
			outbound.TLS = &model.SingBoxTLS{
				Enabled:    true,
				ServerName: node.Host,
				Insecure:   node.SkipCert,
			}
		}

	case "tuic":
		outbound.Type = "tuic"
		outbound.UUID = node.UUID
		outbound.Password = node.Password
		if node.CongestionControl != "" {
			outbound.CongestionControl = node.CongestionControl
		}
		// TUIC 需要 TLS
		if outbound.TLS == nil {
			outbound.TLS = &model.SingBoxTLS{
				Enabled:    true,
				ServerName: node.Host,
				Insecure:   node.SkipCert,
			}
		}

	case "wireguard":
		outbound.Type = "wireguard"
		outbound.PublicKey = node.PublicKey
		outbound.PrivateKey = node.PrivateKey
		outbound.MTU = node.MTU
		// 解析 LocalAddress from ExtraConfig
		if ip := s.getExtraConfigValue(node, "ip"); ip != "" {
			outbound.LocalAddress = append(outbound.LocalAddress, ip)
		}
		if ipv6 := s.getExtraConfigValue(node, "ipv6"); ipv6 != "" {
			outbound.LocalAddress = append(outbound.LocalAddress, ipv6)
		}

	case "socks5", "socks":
		outbound.Type = "socks"
		if node.Username != "" {
			outbound.Username = node.Username
			outbound.Password = node.Password
		}

	case "http":
		outbound.Type = "http"
		if node.Username != "" {
			outbound.Username = node.Username
			outbound.Password = node.Password
		}

	default:
		outbound.Type = node.Type
	}

	return outbound
}

// buildTransport builds transport configuration
func (s *SingBoxConfigService) buildTransport(node *model.Node) *model.SingBoxTransport {
	transport := &model.SingBoxTransport{}

	switch node.Network {
	case "ws":
		transport.Type = "ws"
		transport.Path = node.Path
		if node.Host != "" {
			transport.Headers = map[string]string{
				"Host": node.Host,
			}
		}
		// Early data support
		if node.MaxEarlyData > 0 {
			transport.MaxEarlyData = node.MaxEarlyData
		} else if maxEarlyData := s.getExtraConfigValue(node, "max-early-data"); maxEarlyData != "" {
			fmt.Sscanf(maxEarlyData, "%d", &transport.MaxEarlyData)
		}
		if node.EarlyDataHeader != "" {
			transport.EarlyDataHeaderName = node.EarlyDataHeader
		} else if earlyDataHeader := s.getExtraConfigValue(node, "early-data-header-name"); earlyDataHeader != "" {
			transport.EarlyDataHeaderName = earlyDataHeader
		}

	case "grpc":
		transport.Type = "grpc"
		// 优先使用 ServiceName 字段，其次使用 Path
		if node.ServiceName != "" {
			transport.ServiceName = node.ServiceName
		} else {
			transport.ServiceName = node.Path
		}

	case "h2", "http":
		transport.Type = "http"
		transport.Path = node.Path
		if node.Host != "" {
			transport.Host = node.Host
		}
		transport.Method = "GET" // default method

	default:
		return nil
	}

	return transport
}

// getExtraConfigValue gets a value from ExtraConfig JSON
func (s *SingBoxConfigService) getExtraConfigValue(node *model.Node, key string) string {
	if node.ExtraConfig == "" {
		return ""
	}
	var extra map[string]interface{}
	if err := json.Unmarshal([]byte(node.ExtraConfig), &extra); err != nil {
		return ""
	}
	if val, ok := extra[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case bool:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// buildRoute builds route configuration
func (s *SingBoxConfigService) buildRoute(rules []model.Rule, nodeMap, groupMap map[uint]string, nodeTags map[uint]string, groupOutMap map[string]string) *model.SingBoxRoute {
	route := &model.SingBoxRoute{
		Rules: []model.SingBoxRouteRule{},
	}

	// Add DNS hijack rule first
	route.Rules = append(route.Rules, model.SingBoxRouteRule{
		Type:     "logical",
		Outbound: "dns-out",
		Protocol: "dns",
	})

	// Convert rules
	for _, r := range rules {
		targetOutbound := s.resolveTarget(r, nodeMap, groupMap, nodeTags, groupOutMap)
		if targetOutbound == "" {
			continue
		}

		rule := s.convertRule(&r, targetOutbound)
		if rule != nil {
			route.Rules = append(route.Rules, *rule)
		}
	}

	// Build rule-set entries from RULE-SET rules
	ruleSetMap := make(map[string]bool)
	for _, r := range rules {
		if r.Type == "RULE-SET" && !ruleSetMap[r.Payload] {
			ruleSetMap[r.Payload] = true
		}
	}

	// Add rule-set configurations
	for setName := range ruleSetMap {
		route.RuleSet = append(route.RuleSet, model.SingBoxRuleSet{
			Tag:            setName,
			Type:           "remote",
			Format:         "binary",
			URL:            fmt.Sprintf("https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/%s.srs", setName),
			DownloadDetour: "proxy",
			UpdateInterval: "24h",
		})
	}

	// Final rule - route unmatched traffic to proxy
	route.Final = "proxy"

	return route
}

// resolveTarget resolves rule target to outbound tag
func (s *SingBoxConfigService) resolveTarget(rule model.Rule, nodeMap, groupMap map[uint]string, nodeTags map[uint]string, groupOutMap map[string]string) string {
	// Built-in targets
	switch rule.Target {
	case "DIRECT":
		return "direct"
	case "REJECT", "REJECT-DROP":
		return "block"
	case "PROXY":
		return "proxy"
	}

	// Check if it's a node or group
	switch rule.TargetType {
	case "node":
		// Try to parse as ID
		var nodeID uint
		if _, err := fmt.Sscanf(rule.Target, "%d", &nodeID); err == nil {
			if tag, ok := nodeTags[nodeID]; ok {
				return tag
			}
		}
		// Try by name
		for id, name := range nodeMap {
			if name == rule.Target {
				if tag, ok := nodeTags[id]; ok {
					return tag
				}
			}
		}

	case "group":
		// Try to parse as ID
		var groupID uint
		if _, err := fmt.Sscanf(rule.Target, "%d", &groupID); err == nil {
			if name, ok := groupMap[groupID]; ok {
				if tag, ok := groupOutMap[name]; ok {
					return tag
				}
			}
		}
		// Try by name
		if tag, ok := groupOutMap[rule.Target]; ok {
			return tag
		}

	default:
		// Check if it matches a group name
		if tag, ok := groupOutMap[rule.Target]; ok {
			return tag
		}
		// Check if it matches a node name
		for id, name := range nodeMap {
			if name == rule.Target {
				if tag, ok := nodeTags[id]; ok {
					return tag
				}
			}
		}
	}

	return ""
}

// convertRule converts Clash rule to sing-box route rule
func (s *SingBoxConfigService) convertRule(rule *model.Rule, outbound string) *model.SingBoxRouteRule {
	if rule == nil || outbound == "" {
		return nil
	}

	sboxRule := &model.SingBoxRouteRule{
		Outbound: outbound,
	}

	switch rule.Type {
	case "DOMAIN":
		sboxRule.Domain = []string{rule.Payload}

	case "DOMAIN-SUFFIX":
		sboxRule.DomainSuffix = []string{rule.Payload}

	case "DOMAIN-KEYWORD":
		sboxRule.DomainKeyword = []string{rule.Payload}

	case "IP-CIDR":
		payload := rule.Payload
		if !strings.Contains(payload, "/") {
			payload = payload + "/32"
		}
		sboxRule.IPCIDR = []string{payload}

	case "SRC-IP-CIDR":
		payload := rule.Payload
		if !strings.Contains(payload, "/") {
			payload = payload + "/32"
		}
		sboxRule.SourceIPCIDR = []string{payload}

	case "GEOIP":
		sboxRule.GeoIP = []string{rule.Payload}

	case "SRC-GEOIP":
		sboxRule.SourceGeoIP = []string{rule.Payload}

	case "GEOSITE":
		sboxRule.Geosite = []string{rule.Payload}

	case "DST-PORT":
		var port int
		fmt.Sscanf(rule.Payload, "%d", &port)
		sboxRule.Port = []int{port}

	case "SRC-PORT":
		var port int
		fmt.Sscanf(rule.Payload, "%d", &port)
		sboxRule.SourcePort = []int{port}

	case "RULE-SET":
		sboxRule.RuleSet = []string{rule.Payload}

	case "PROCESS-NAME":
		sboxRule.ProcessName = []string{rule.Payload}

	case "NETWORK":
		sboxRule.Network = rule.Payload // "tcp" or "udp"

	case "MATCH":
		// MATCH is handled by route.Final
		return nil

	default:
		// Unsupported rule type
		return nil
	}

	return sboxRule
}

// sanitizeTag sanitizes a name to be used as a tag
func sanitizeTag(name string) string {
	// Replace spaces and special characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
