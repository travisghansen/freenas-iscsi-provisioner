package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &Portal{}
)

type Portal struct {
	Id                  int      `json:"id,omitempty"`
	Tag                 int      `json:"iscsi_target_portal_tag,omitempty"`
	Comment             string   `json:"iscsi_target_portal_comment,omitempty"`
	Discoveryauthgroup  string   `json:"iscsi_target_portal_discoveryauthgroup,omitempty"`
	Discoveryauthmethod string   `json:"iscsi_target_portal_discoveryauthmethod,omitempty"`
	Ips                 []string `json:"iscsi_target_portal_ips,omitempty"`
}

func (p *Portal) CopyFrom(source FreenasResource) error {
	src, ok := source.(*Portal)
	if ok {
		p.Id = src.Id
		p.Tag = src.Tag
		p.Comment = src.Comment
		p.Discoveryauthgroup = src.Discoveryauthgroup
		p.Discoveryauthmethod = src.Discoveryauthmethod
		p.Ips = src.Ips
	}

	return errors.New("Cannot copy, src is not a Portal")
}

func (p *Portal) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/portal/%d/", p.Id)
	var portal Portal
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&portal)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	p.CopyFrom(&portal)

	return nil
}

func (p *Portal) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/portal/"
	var portal Portal
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(p).Receive(&portal, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating portal for %+v - %v", *p, body))
	}

	p.CopyFrom(&portal)

	return nil
}

func (p *Portal) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/portal/%d/", p.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
