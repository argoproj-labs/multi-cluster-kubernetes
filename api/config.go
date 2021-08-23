package api

import (
	"k8s.io/client-go/tools/clientcmd/api"
	"time"

	"k8s.io/client-go/rest"
)

func NewConfig(c rest.Config, authInfo api.AuthInfo) Config {
	return Config{
		Host:               c.Host,
		APIPath:            c.APIPath,
		Username:           authInfo.Username,
		Password:           authInfo.Password,
		BearerToken:        authInfo.Token,
		TLSClientConfig:    TLSClientConfig(c.TLSClientConfig),
		UserAgent:          c.UserAgent,
		DisableCompression: c.DisableCompression,
		QPS:                c.QPS,
		Burst:              c.Burst,
		Timeout:            c.Timeout,
	}
}

type Config struct {
	Host               string                   `json:"host,omitempty"`
	APIPath            string                   `json:"apiPath,omitempty"`
	Username           string                   `json:"username,omitempty"`
	Password           string                   `json:"password,omitempty"`
	BearerToken        string                   `json:"bearerToken,omitempty"`
	TLSClientConfig    TLSClientConfig          `json:"tlsClient_config"`
	UserAgent          string                   `json:"userAgent,omitempty"`
	DisableCompression bool                     `json:"disableCompression,omitempty"`
	QPS                float32                  `json:"qps,omitempty"`
	Burst              int                      `json:"burst,omitempty"`
	Timeout            time.Duration            `json:"timeout,omitempty"`
	ExecProvider       *api.ExecConfig          `json:"execProvider,omitempty"`
	AuthProvider       *api.AuthProviderConfig  `json:"authProvider,omitempty"`
	Impersonate        rest.ImpersonationConfig `json:"impersonate,omitempty"`
}

type TLSClientConfig struct {
	Insecure   bool     `json:"insecure,omitempty"`
	ServerName string   `json:"serverName,omitempty"`
	CertFile   string   `json:"certFile,omitempty"`
	KeyFile    string   `json:"keyFile,omitempty"`
	CAFile     string   `json:"caFile,omitempty"`
	CertData   []byte   `json:"certData,omitempty"`
	KeyData    []byte   `datapolicy:"security-key" json:"keyData,omitempty"`
	CAData     []byte   `json:"caData,omitempty"`
	NextProtos []string `json:"nextProtos,omitempty"`
}

func (c Config) RestConfig() *rest.Config {
	return &rest.Config{
		Host:               c.Host,
		APIPath:            c.APIPath,
		Username:           c.Username,
		Password:           c.Password,
		BearerToken:        c.BearerToken,
		Impersonate:        c.Impersonate,
		AuthProvider:       c.AuthProvider,
		ExecProvider:       c.ExecProvider,
		TLSClientConfig:    rest.TLSClientConfig(c.TLSClientConfig),
		UserAgent:          c.UserAgent,
		DisableCompression: c.DisableCompression,
		QPS:                c.QPS,
		Burst:              c.Burst,
		Timeout:            c.Timeout,
	}
}
