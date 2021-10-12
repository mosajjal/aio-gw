package conf

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type WebserverSettings struct {
	Enabled        bool   `yaml:"enabled" inputtype:"toggle" default:"true"`
	Bind           string `yaml:"bind"    inputtype:"text"   default:":8080"`
	Authentication struct {
		Enabled  bool   `yaml:"enabled"  inputtype:"toggle" default:"true"`
		Username string `yaml:"username" inputtype:"text"   default:"admin"`
		Password string `yaml:"password" inputtype:"text"   default:"admin"`
	} `yaml:"authentication"`
	Tls struct {
		Enabled bool   `yaml:"enabled" inputtype:"toggle" default:"false"`
		Cert    string `yaml:"cert"    inputtype:"file"   default:""`
		Key     string `yaml:"key"     inputtype:"file"   default:""`
	} `yaml:"tls"`
	Logging struct {
		Enabled bool   `yaml:"enabled" inputtype:"toggle"  default:"true"`
		Level   string `yaml:"level"   inputtype:"options" options:"debug,info,warning" default:"debug"`
	} `yaml:"logging"`
}

type UpstreamSettings struct {
	Enabled         bool   `yaml:"enabled" inputtype:"toggle"  default:"true"`
	Method          string `yaml:"method"  inputtype:"options" options:"sinkhole,dummy,nat,tls_decryption,openvpn,wireguard,tor,socks,http" default:"tls_decryption"`
	SinkholeOptions struct {
	} `yaml:"sinkhole_options"`
	DummyOptions struct {
	} `yaml:"dummy_options"`
	TlsDecrpytionOptions struct {
		Ports  []int `yaml:"ports" inputtype:"text" default:"443,8443,8444,853,990,465,993,995,5061"`
		Target struct {
			Ip   string `yaml:"ip" inputtype:"text" default:"127.0.0.1"`
			Port int    `yaml:"port" inputtype:"text" default:"10443"`
		} `yaml:"nat_target"`
	} `yaml:"tls_decryption_options"`
	OpenvpnOptions struct {
		// TODO: add options
	} `yaml:"openvpn_options"`
	WireguardOptions struct {
		// TODO: add options
	} `yaml:"wireguard_options"`
	TorOptions struct {
		// TODO: add options
	} `yaml:"tor_options"`
	SocksOptions struct {
		// TODO: add options
	} `yaml:"socks_options"`
	HttpOptions struct {
		// TODO: add options
	} `yaml:"http_options"`
	NatOptiones struct {
		// TODO: add options
	} `yaml:"nat_options"`
}

type ProxySettings struct {
	Enabled        bool   `yaml:"enabled" inputtype:"toggle" default:"true"`
	Bind           string `yaml:"bind"    inputtype:"text"   default:":8081"`
	Authentication struct {
		Enabled  bool   `yaml:"enabled"  inputtype:"toggle" default:"true"`
		Username string `yaml:"username" inputtype:"text"   default:"admin"`
		Password string `yaml:"password" inputtype:"text"   default:"admin"`
	} `yaml:"authentication"`
	Type string `yaml:"type"`
}

type ServiceSettings struct {
	Containers []containerSettings
}

// simple podman container settings
type containerSettings struct {
	Enabled        bool     `yaml:"enabled"       inputtype:"toggle" default:"true"`
	Name           string   `yaml:"name"          inputtype:"text"   default:""`
	EnabledAtStart bool     `yaml:"enabled_at_start" inputtype:"toggle" default:"true"`
	StartDelay     int      `yaml:"start_delay"    inputtype:"text"   default:"5"`
	PodmanOptions  []string `yaml:"options" inputtype:"text" default:""`
	EntryOptions   []string `yaml:"entry_options" inputtype:"text" default:""`
	Image          string   `yaml:"image" inputtype:"text" default:"hello-world"`
}

// default settings for containers
var elasticSettings = containerSettings{
	Enabled:        true,
	Name:           "elasticsearch",
	EnabledAtStart: true,
	StartDelay:     5,
	PodmanOptions:  []string{"-v", "/data/es-data:/usr/share/elasticsearch/data", "-e", "bootstrap.memory_lock=true", "-e", "\"ES_JAVA_OPTS=-Xms512m -Xmx512m\"", "-e", "discovery.type=single-node"},
	EntryOptions:   nil,
	Image:          "docker.elastic.co/elasticsearch/elasticsearch:7.13.2",
}
var arkimeSettings = containerSettings{
	Enabled:        true,
	Name:           "arkime",
	EnabledAtStart: true,
	StartDelay:     5,
	EntryOptions:   []string{"--createAdminUser=true", "--configPath=/opt/arkime/etc/config.ini"},
	PodmanOptions:  []string{"-v", "/data/moloch/raw:/opt/arkime/raw", "-v", "/data/moloch/arkime.ini:/opt/arkime/etc/config.ini", "--ulimit", "memlock=-1:-1"},
	Image:          "mosajjal/arkime:3.1.0",
}
var polarproxySettings = containerSettings{
	Enabled:        true,
	Name:           "polarproxy",
	EnabledAtStart: true,
	StartDelay:     5,
	EntryOptions:   []string{"-v", "-p", "10443,80,443", "--certhttp", "1081", "--pcapoveripconnect", "127.0.0.1:57012", "--cn", "Fortinet_CA_SSL"},
	PodmanOptions:  nil,
	Image:          "mosajjal/polarproxy:latest",
}

var defaultServiceSettings = ServiceSettings{
	Containers: []containerSettings{
		elasticSettings,
		arkimeSettings,
		polarproxySettings,
	},
}

// used to write and read config file
type FullSettings struct {
	WebserverSettings WebserverSettings `yaml:"webserver_settings"`
	UpstreamSettings  UpstreamSettings  `yaml:"upstream_settings"`
	ProxySettings     ProxySettings     `yaml:"proxy_settings"`
	ServiceSettings   ServiceSettings   `yaml:"service_settings"`
}

var GlobalWebserverSettings WebserverSettings
var GlobalUpstreamSettings UpstreamSettings
var GlobalProxySettings ProxySettings
var GlobalServiceSettings ServiceSettings

func GenerateDefaultSettings() {
	Defaults := FullSettings{
		WebserverSettings: WebserverSettings{},
		UpstreamSettings:  UpstreamSettings{},
		ProxySettings:     ProxySettings{},
		ServiceSettings:   ServiceSettings{},
	}
	err := envconfig.Process("aio-gw", &Defaults)
	// default confing is different: TODO
	Defaults.ServiceSettings = defaultServiceSettings
	if err != nil {
		log.Fatal(err.Error())
	}

	confYaml, _ := yaml.Marshal(Defaults)
	fmt.Println(string(confYaml))
}

func LoadSettingsFromFile(filename string) error {
	FileSettings := FullSettings{}
	OpenFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(OpenFile, &FileSettings)
	if err != nil {
		return err
	}
	GlobalWebserverSettings = FileSettings.WebserverSettings
	GlobalProxySettings = FileSettings.ProxySettings
	GlobalServiceSettings = FileSettings.ServiceSettings
	GlobalUpstreamSettings = FileSettings.UpstreamSettings
	log.Infof("Loaded settings from file: " + filename)
	return nil
}
