package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &TargetToExtent{}
)

type TargetToExtent struct {
	Id     int  `json:"id,omitempty"`
	Extent int  `json:"iscsi_extent,omitempty"`
	Lunid  *int `json:"iscsi_lunid,omitempty"`
	Target int  `json:"iscsi_target,omitempty"`
}

func (t *TargetToExtent) CopyFrom(source FreenasResource) error {
	src, ok := source.(*TargetToExtent)
	if ok {
		t.Id = src.Id
		t.Extent = src.Extent
		t.Lunid = src.Lunid
		t.Target = src.Target
	}

	return errors.New("Cannot copy, src is not a TargetToExtent")
}

func (t *TargetToExtent) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targettoextent/%d/", t.Id)
	var targetToExtent TargetToExtent
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&targetToExtent)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	t.CopyFrom(&targetToExtent)

	return nil
}

func (t *TargetToExtent) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/targettoextent/"
	var targetToExtent TargetToExtent
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&targetToExtent, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating targettoextent for %+v - %v", *t, body))
	}

	t.CopyFrom(&targetToExtent)

	return nil
}

func (t *TargetToExtent) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targettoextent/%d/", t.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
