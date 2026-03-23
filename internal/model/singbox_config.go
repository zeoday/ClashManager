package model

// SingBoxConfig represents the root of sing-box configuration
type SingBoxConfig struct {
	Log       *SingBoxLog       `json:"log,omitempty"`
	DNS       *SingBoxDNS       `json:"dns,omitempty"`
	Inbounds  []SingBoxInbound  `json:"inbounds"`
	Outbounds []SingBoxOutbound `json:"outbounds"`
	Route     *SingBoxRoute     `json:"route,omitempty"`
}

// SingBoxLog represents log configuration
type SingBoxLog struct {
	Disabled  bool   `json:"disabled,omitempty"`
	Level     string `json:"level,omitempty"`
	Output    string `json:"output,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

// SingBoxDNS represents DNS configuration
type SingBoxDNS struct {
	Servers       []SingBoxDNSServer `json:"servers,omitempty"`
	Rules         []SingBoxDNSRule   `json:"rules,omitempty"`
	Final         string             `json:"final,omitempty"`
	Strategy      string             `json:"strategy,omitempty"`
	DisableCache  bool               `json:"disable_cache,omitempty"`
	DisableExpire bool               `json:"disable_expire,omitempty"`
}

// SingBoxDNSServer represents a DNS server
type SingBoxDNSServer struct {
	Tag             string `json:"tag"`
	Address         string `json:"address"`
	Detour          string `json:"detour,omitempty"`
	AddressResolver string `json:"address_resolver,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	AddressStrategy string `json:"address_strategy,omitempty"`
	DisableCache    bool   `json:"disable_cache,omitempty"`
	PreferH3        bool   `json:"prefer_h3,omitempty"`
}

// SingBoxDNSRule represents a DNS rule
type SingBoxDNSRule struct {
	Outbound      string   `json:"outbound,omitempty"`
	Server        string   `json:"server,omitempty"`
	DisableCache  bool     `json:"disable_cache,omitempty"`
	Rules         []string `json:"rule,omitempty"`
	Domain        []string `json:"domain,omitempty"`
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	Geosite       []string `json:"geosite,omitempty"`
}

// SingBoxInbound represents inbound configuration
type SingBoxInbound struct {
	Type                     string                 `json:"type"`
	Tag                      string                 `json:"tag"`
	Listen                   string                 `json:"listen,omitempty"`
	ListenPort               int                    `json:"listen_port,omitempty"`
	TCPFastOpen              bool                   `json:"tcp_fast_open,omitempty"`
	UDPFragment              bool                   `json:"udp_fragment,omitempty"`
	Sniff                    bool                   `json:"sniff,omitempty"`
	SniffOverrideDestination bool                   `json:"sniff_override_destination,omitempty"`
	DomainStrategy           string                 `json:"domain_strategy,omitempty"`
	SetSystemProxy           bool                   `json:"set_system_proxy,omitempty"`
	Options                  map[string]interface{} `json:"options,omitempty"`
}

// SingBoxOutbound represents outbound configuration
type SingBoxOutbound struct {
	Type        string            `json:"type"`
	Tag         string            `json:"tag"`
	Server      string            `json:"server,omitempty"`
	ServerPort  int               `json:"server_port,omitempty"`
	Method      string            `json:"method,omitempty"`
	Password    string            `json:"password,omitempty"`
	UUID        string            `json:"uuid,omitempty"`
	Flow        string            `json:"flow,omitempty"`
	AlterId     int               `json:"alter_id,omitempty"`
	Security    string            `json:"security,omitempty"`
	TLS         *SingBoxTLS       `json:"tls,omitempty"`
	Transport   *SingBoxTransport `json:"transport,omitempty"`
	Multiplex   *SingBoxMultiplex `json:"multiplex,omitempty"`
	Detour      string            `json:"detour,omitempty"`
	BrutalDebug bool              `json:"brutal_debug,omitempty"`
	// Selector/URLTest group fields
	Outbounds []string `json:"outbounds,omitempty"`
	Default   string   `json:"default,omitempty"`
	URL       string   `json:"url,omitempty"`
	Interval  int      `json:"interval,omitempty"`
	Tolerance int      `json:"tolerance,omitempty"`
	// SOCKS5/HTTP 认证
	Username string `json:"username,omitempty"`
	// Hysteria v1
	AuthString string `json:"auth_str,omitempty"`
	// Hysteria2 带宽
	UpMbps   int `json:"up_mbps,omitempty"`
	DownMbps int `json:"down_mbps,omitempty"`
	// Brutal 拥塞控制
	Brutal *SingBoxBrutal `json:"brutal,omitempty"`
	// TUIC
	CongestionControl string `json:"congestion_control,omitempty"`
	// WireGuard
	PublicKey     string   `json:"public_key,omitempty"`
	PrivateKey    string   `json:"private_key,omitempty"`
	LocalAddress  []string `json:"local_address,omitempty"`
	MTU           int      `json:"mtu,omitempty"`
	PeerPublicKey string   `json:"peer_public_key,omitempty"`
	PreSharedKey  string   `json:"pre_shared_key,omitempty"`
	Reserved      []int    `json:"reserved,omitempty"`
}

// SingBoxTLS represents TLS configuration
type SingBoxTLS struct {
	Enabled         bool            `json:"enabled"`
	ServerName      string          `json:"server_name,omitempty"`
	Insecure        bool            `json:"insecure,omitempty"`
	ALPN            []string        `json:"alpn,omitempty"`
	MinVersion      string          `json:"min_version,omitempty"`
	MaxVersion      string          `json:"max_version,omitempty"`
	CertificatePath string          `json:"certificate_path,omitempty"`
	Certificate     string          `json:"certificate,omitempty"`
	ECH             *SingBoxECH     `json:"ech,omitempty"`
	UTLS            *SingBoxUTLS    `json:"utls,omitempty"`
	Reality         *SingBoxReality `json:"reality,omitempty"`
}

// SingBoxECH represents ECH configuration
type SingBoxECH struct {
	Enabled bool     `json:"enabled"`
	Config  []string `json:"config,omitempty"`
}

// SingBoxUTLS represents UTLS configuration
type SingBoxUTLS struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// SingBoxReality represents Reality configuration
type SingBoxReality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
	ShortId   string `json:"short_id,omitempty"`
}

// SingBoxTransport represents transport configuration
type SingBoxTransport struct {
	Type                string            `json:"type"`
	Path                string            `json:"path,omitempty"`
	Host                string            `json:"host,omitempty"`
	Method              string            `json:"method,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	ServiceName         string            `json:"service_name,omitempty"`
	IdleTimeout         string            `json:"idle_timeout,omitempty"`
	PingTimeout         string            `json:"ping_timeout,omitempty"`
	MaxEarlyData        int               `json:"max_early_data,omitempty"`
	EarlyDataHeaderName string            `json:"early_data_header_name,omitempty"`
}

// SingBoxMultiplex represents multiplex configuration
type SingBoxMultiplex struct {
	Enabled        bool           `json:"enabled"`
	Protocol       string         `json:"protocol,omitempty"`
	MaxConnections int            `json:"max_connections,omitempty"`
	MinStreams     int            `json:"min_streams,omitempty"`
	MaxStreams     int            `json:"max_streams,omitempty"`
	Padding        bool           `json:"padding,omitempty"`
	Brutal         *SingBoxBrutal `json:"brutal,omitempty"`
}

// SingBoxBrutal represents brutal congestion control
type SingBoxBrutal struct {
	Enabled bool   `json:"enabled"`
	Up      string `json:"up,omitempty"`
	Down    string `json:"down,omitempty"`
}

// SingBoxRoute represents route configuration
type SingBoxRoute struct {
	Rules       []SingBoxRouteRule `json:"rules,omitempty"`
	RuleSet     []SingBoxRuleSet   `json:"rule_set,omitempty"`
	Final       string             `json:"final,omitempty"`
	FindProcess bool               `json:"find_process,omitempty"`
	AutoDetect  bool               `json:"auto_detect_interface,omitempty"`
	DefaultMark int                `json:"default_mark,omitempty"`
}

// SingBoxRouteRule represents a routing rule
type SingBoxRouteRule struct {
	Type            string   `json:"type,omitempty"`
	Outbound        string   `json:"outbound,omitempty"`
	Inbound         []string `json:"inbound,omitempty"`
	Network         string   `json:"network,omitempty"`
	Protocol        string   `json:"protocol,omitempty"`
	Domain          []string `json:"domain,omitempty"`
	DomainSuffix    []string `json:"domain_suffix,omitempty"`
	DomainKeyword   []string `json:"domain_keyword,omitempty"`
	DomainRegex     []string `json:"domain_regex,omitempty"`
	Geosite         []string `json:"geosite,omitempty"`
	SourceGeoIP     []string `json:"source_geoip,omitempty"`
	GeoIP           []string `json:"geoip,omitempty"`
	SourceIPCIDR    []string `json:"source_ip_cidr,omitempty"`
	IPCIDR          []string `json:"ip_cidr,omitempty"`
	SourcePort      []int    `json:"source_port,omitempty"`
	SourcePortRange []string `json:"source_port_range,omitempty"`
	Port            []int    `json:"port,omitempty"`
	PortRange       []string `json:"port_range,omitempty"`
	ProcessName     []string `json:"process_name,omitempty"`
	ProcessPath     []string `json:"process_path,omitempty"`
	RuleSet         []string `json:"rule_set,omitempty"`
	Invert          bool     `json:"invert,omitempty"`
}

// SingBoxRuleSet represents a rule-set
type SingBoxRuleSet struct {
	Tag            string   `json:"tag"`
	Type           string   `json:"type"`
	Format         string   `json:"format,omitempty"`
	URL            string   `json:"url,omitempty"`
	Path           string   `json:"path,omitempty"`
	DownloadDetour string   `json:"download_detour,omitempty"`
	UpdateInterval string   `json:"update_interval,omitempty"`
	Exclude        []string `json:"exclude,omitempty"`
	Include        []string `json:"include,omitempty"`
}

// SingBoxWireGuard represents WireGuard outbound configuration
type SingBoxWireGuard struct {
	Type          string   `json:"type"`
	Tag           string   `json:"tag"`
	Server        string   `json:"server,omitempty"`
	ServerPort    int      `json:"server_port,omitempty"`
	LocalAddress  []string `json:"local_address"`
	PrivateKey    string   `json:"private_key"`
	PeerPublicKey string   `json:"peer_public_key,omitempty"`
	PreSharedKey  string   `json:"pre_shared_key,omitempty"`
	Reserved      []int    `json:"reserved,omitempty"`
	MTU           int      `json:"mtu,omitempty"`
}

// SingBoxTUIC represents TUIC outbound configuration
type SingBoxTUIC struct {
	Type              string      `json:"type"`
	Tag               string      `json:"tag"`
	Server            string      `json:"server"`
	ServerPort        int         `json:"server_port"`
	UUID              string      `json:"uuid"`
	Password          string      `json:"password"`
	CongestionControl string      `json:"congestion_control,omitempty"`
	ALPN              []string    `json:"alpn,omitempty"`
	Heartbeat         string      `json:"heartbeat,omitempty"`
	UDPRelayMode      string      `json:"udp_relay_mode,omitempty"`
	ZeroRTTHandshake  bool        `json:"zero_rtt_handshake,omitempty"`
	TLS               *SingBoxTLS `json:"tls,omitempty"`
}

// SingBoxHysteria represents Hysteria v1 outbound configuration
type SingBoxHysteria struct {
	Type       string      `json:"type"`
	Tag        string      `json:"tag"`
	Server     string      `json:"server"`
	ServerPort int         `json:"server_port"`
	AuthString string      `json:"auth_str,omitempty"`
	UpMbps     int         `json:"up_mbps,omitempty"`
	DownMbps   int         `json:"down_mbps,omitempty"`
	Obfs       string      `json:"obfs,omitempty"`
	TLS        *SingBoxTLS `json:"tls,omitempty"`
}

// SingBoxTunInbound represents TUN inbound configuration
type SingBoxTunInbound struct {
	Type                   string   `json:"type"`
	Tag                    string   `json:"tag"`
	InterfaceName          string   `json:"interface_name,omitempty"`
	MTU                    int      `json:"mtu,omitempty"`
	Inet4Address           string   `json:"inet4_address,omitempty"`
	Inet6Address           string   `json:"inet6_address,omitempty"`
	AutoRoute              bool     `json:"auto_route,omitempty"`
	StrictRoute            bool     `json:"strict_route,omitempty"`
	Stack                  string   `json:"stack,omitempty"`
	IncludePackage         []string `json:"include_package,omitempty"`
	ExcludePackage         []string `json:"exclude_package,omitempty"`
	IncludeAndroidUser     []int    `json:"include_android_user,omitempty"`
	IncludeAndroidPackage  []string `json:"include_android_package,omitempty"`
	ExcludeAndroidPackage  []string `json:"exclude_android_package,omitempty"`
	EndpointIndependentNat bool     `json:"endpoint_independent_nat,omitempty"`
	UDPTimeout             string   `json:"udp_timeout,omitempty"`
}
