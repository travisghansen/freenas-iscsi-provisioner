package freenas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &TargetGroup{}
)

// TargetGroup represents a TargetGroup instance
type TargetGroup struct {
	ID             int    `json:"id,omitempty"`
	Target         int    `json:"iscsi_target,omitempty"`
	Authgroup      int    `json:"iscsi_target_authgroup,omitempty"`
	Authtype       string `json:"iscsi_target_authtype,omitempty"`
	Initialdigest  string `json:"iscsi_target_initialdigest,omitempty"`
	Initiatorgroup int    `json:"iscsi_target_initiatorgroup,omitempty"`
	Portalgroup    int    `json:"iscsi_target_portalgroup,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (t *TargetGroup) CopyFrom(source Resource) error {
	src, ok := source.(*TargetGroup)
	if ok {
		t.ID = src.ID
		t.Target = src.Target
		t.Authgroup = src.Authgroup
		t.Authtype = src.Authtype
		t.Initialdigest = src.Initialdigest
		t.Initiatorgroup = src.Initiatorgroup
		t.Portalgroup = src.Portalgroup
	}

	return errors.New("Cannot copy, src is not a TargetGroup")
}

// Get gets a TargetGroup instance
func (t *TargetGroup) Get(server *Server) (*http.Response, error) {
	// find by ID
	if t.ID > 0 {
		endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targetgroup/%d/", t.ID)
		var targetGroup TargetGroup
		resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&targetGroup)
		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		t.CopyFrom(&targetGroup)

		return resp, nil
	}

	// find by portal group ID and target ID
	if t.Portalgroup > 0 && t.Target > 0 {
		endpoint := "/api/v1.0/services/iscsi/targetgroup/?limit=1000"
		var list []TargetGroup
		var e interface{}
		resp, err := server.getSlingConnection().Get(endpoint).Receive(&list, &e)

		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		if resp.StatusCode != 200 {
			body, _ := json.Marshal(e)
			return resp, fmt.Errorf("Error getting TargetGroup target ID: %d, portal group ID: %d - message: %v, status: %d", t.Target, t.Portalgroup, string(body), resp.StatusCode)
		}

		for _, item := range list {
			if item.Target == t.Target && item.Portalgroup == t.Portalgroup {
				t.CopyFrom(&item)
				glog.Infof("found TargetGroup target ID: %d, portal group ID: %d - %+v", t.Target, t.Portalgroup, *t)
				return resp, nil
			}
		}
	}

	// Nothing found
	return nil, errors.New("no Target has been found")
}

// Create creates a TargetGroup instance
func (t *TargetGroup) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/targetgroup/"
	var targetGroup TargetGroup
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&targetGroup, nil)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating targetgroup for %+v - %v", *t, body)
	}

	t.CopyFrom(&targetGroup)

	return resp, nil
}

// Delete deletes a TargetGroup instance
func (t *TargetGroup) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targetgroup/%d/", t.ID)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return resp, err
}
