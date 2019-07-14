package freenas

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

// Resource basic interface for http interactions with various FreeNAS resources
type Resource interface {
	CopyFrom(source Resource) error
	Get(server *Server) (*http.Response, error)
	Create(server *Server) (*http.Response, error)
	Delete(server *Server) (*http.Response, error)
}

// ErrorResponse generic object to contain FreeNAS API errors
type ErrorResponse struct {
	DefaultAll []string `json:"__all__,omitempty"`
}

// Server struct representing connection details
type Server struct {
	Protocol                 string
	Host, Username, Password string
	Port                     int
	InsecureSkipVerify       bool
	url                      string
}

// NewFreenasServer gets a new connection instance
func NewFreenasServer(protocol string, host string, port int, username, password string, insecure bool) *Server {
	u := fmt.Sprintf("%s://%s:%d", protocol, host, port)
	return &Server{
		Protocol:           protocol,
		Host:               host,
		Port:               port,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
		url:                u,
	}
}

func (s *Server) getSlingConnection() *sling.Sling {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: s.InsecureSkipVerify},
	}

	if s.Protocol == "http" {
		tr.TLSClientConfig = nil
	}

	httpClient := &http.Client{Transport: tr}
	return sling.New().Client(httpClient).Base(s.url).SetBasicAuth(s.Username, s.Password)
}
