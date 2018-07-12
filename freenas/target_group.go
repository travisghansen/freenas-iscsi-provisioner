package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &TargetGroup{}
)

type TargetGroup struct {
	Id             int    `json:"id,omitempty"`
	Target         int    `json:"iscsi_target,omitempty"`
	Authgroup      int    `json:"iscsi_target_authgroup,omitempty"`
	Authtype       string `json:"iscsi_target_authtype,omitempty"`
	Initialdigest  string `json:"iscsi_target_initialdigest,omitempty"`
	Initiatorgroup int    `json:"iscsi_target_initiatorgroup,omitempty"`
	Portalgroup    int    `json:"iscsi_target_portalgroup,omitempty"`
}

func (t *TargetGroup) CopyFrom(source FreenasResource) error {
	src, ok := source.(*TargetGroup)
	if ok {
		t.Id = src.Id
		t.Target = src.Target
		t.Authgroup = src.Authgroup
		t.Authtype = src.Authtype
		t.Initialdigest = src.Initialdigest
		t.Initiatorgroup = src.Initiatorgroup
		t.Portalgroup = src.Portalgroup
	}

	return errors.New("Cannot copy, src is not a TargetGroup")
}

func (t *TargetGroup) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targetgroup/%d/", t.Id)
	var targetGroup TargetGroup
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&targetGroup)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	t.CopyFrom(&targetGroup)

	return nil
}

func (t *TargetGroup) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/targetgroup/"
	var targetGroup TargetGroup
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&targetGroup, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating targetgroup for %+v - %v", *t, body))
	}

	t.CopyFrom(&targetGroup)

	return nil
}

func (t *TargetGroup) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targetgroup/%d/", t.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
