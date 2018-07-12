package freenas

import (
	"errors"
	"github.com/golang/glog"
)

var (
	_ FreenasResource = &ISCSIConfig{}
)

type ISCSIConfig struct {
	Id                 int    `json:"id,omitempty"`
	Basename           string `json:"iscsi_basename,omitempty"`
	ISNSServers        string `json:"iscsi_isns_servers,omitempty"`
	PoolAvailThreshold int    `json:"iscsi_pool_avail_threshold,omitempty"`
}

func (i *ISCSIConfig) CopyFrom(source FreenasResource) error {
	src, ok := source.(*ISCSIConfig)
	if ok {
		i.Id = src.Id
		i.Basename = src.Basename
		i.ISNSServers = src.ISNSServers
		i.PoolAvailThreshold = src.PoolAvailThreshold
	}

	return errors.New("Cannot copy, src is not a ISCSIConfig")
}

func (i *ISCSIConfig) Get(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/globalconfiguration/"
	var iscsiConfig ISCSIConfig
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&iscsiConfig)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	i.CopyFrom(&iscsiConfig)

	return nil
}

func (i *ISCSIConfig) Create(server *FreenasServer) error {
	return errors.New("Create method unavailable")
}

func (i *ISCSIConfig) Delete(server *FreenasServer) error {
	return errors.New("Delete method unavailable")
}
