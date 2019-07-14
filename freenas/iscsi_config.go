package freenas

import (
	"errors"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &ISCSIConfig{}
)

// ISCSIConfig represents an ISCSIConfig instance
type ISCSIConfig struct {
	ID                 int    `json:"id,omitempty"`
	Basename           string `json:"iscsi_basename,omitempty"`
	ISNSServers        string `json:"iscsi_isns_servers,omitempty"`
	PoolAvailThreshold int    `json:"iscsi_pool_avail_threshold,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (i *ISCSIConfig) CopyFrom(source Resource) error {
	src, ok := source.(*ISCSIConfig)
	if ok {
		i.ID = src.ID
		i.Basename = src.Basename
		i.ISNSServers = src.ISNSServers
		i.PoolAvailThreshold = src.PoolAvailThreshold
	}

	return errors.New("Cannot copy, src is not a ISCSIConfig")
}

// Get gets an ISCSIConfig instance
func (i *ISCSIConfig) Get(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/globalconfiguration/"
	var iscsiConfig ISCSIConfig
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&iscsiConfig)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	i.CopyFrom(&iscsiConfig)

	return resp, nil
}

// Create creates an ISCSIConfig instance
func (i *ISCSIConfig) Create(server *Server) (*http.Response, error) {
	return nil, errors.New("Create method unavailable")
}

// Delete deletes an ISCSIConfig instance
func (i *ISCSIConfig) Delete(server *Server) (*http.Response, error) {
	return nil, errors.New("Delete method unavailable")
}
