package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &Target{}
)

type Target struct {
	Id    int    `json:"id,omitempty"`
	Name  string `json:"iscsi_target_name,omitempty"`
	Alias string `json:"iscsi_target_alias,omitempty"`
	Mode  string `json:"iscsi_target_mode,omitempty"`
}

func (t *Target) CopyFrom(source FreenasResource) error {
	src, ok := source.(*Target)
	if ok {
		t.Id = src.Id
		t.Name = src.Name
		t.Alias = src.Alias
		t.Mode = src.Mode
	}

	return errors.New("Cannot copy, src is not a Target")
}

func (t *Target) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/target/%d/", t.Id)
	var target Target
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&target)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	t.CopyFrom(&target)

	return nil
}

func (t *Target) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/target/"
	var target Target
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&target, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating target for %+v - %v", *t, body))
	}

	t.CopyFrom(&target)

	return nil
}

func (t *Target) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/target/%d/", t.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
