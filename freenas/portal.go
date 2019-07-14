package freenas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &Portal{}
)

// Portal represents a Portal instance
type Portal struct {
	ID                  int      `json:"id,omitempty"`
	Tag                 int      `json:"iscsi_target_portal_tag,omitempty"`
	Comment             string   `json:"iscsi_target_portal_comment,omitempty"`
	Discoveryauthgroup  string   `json:"iscsi_target_portal_discoveryauthgroup,omitempty"`
	Discoveryauthmethod string   `json:"iscsi_target_portal_discoveryauthmethod,omitempty"`
	Ips                 []string `json:"iscsi_target_portal_ips,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (p *Portal) CopyFrom(source Resource) error {
	src, ok := source.(*Portal)
	if ok {
		p.ID = src.ID
		p.Tag = src.Tag
		p.Comment = src.Comment
		p.Discoveryauthgroup = src.Discoveryauthgroup
		p.Discoveryauthmethod = src.Discoveryauthmethod
		p.Ips = src.Ips
	}

	return errors.New("Cannot copy, src is not a Portal")
}

// Get gets an Portal instance
func (p *Portal) Get(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/portal/%d/", p.ID)
	var portal Portal
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&portal)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	p.CopyFrom(&portal)

	return resp, nil
}

// Create creates an Portal instance
func (p *Portal) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/portal/"
	var portal Portal
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(p).Receive(&portal, nil)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating portal for %+v - %v", *p, body)
	}

	p.CopyFrom(&portal)

	return resp, nil
}

// Delete deletes an Portal instance
func (p *Portal) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/portal/%d/", p.ID)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return resp, err
}
