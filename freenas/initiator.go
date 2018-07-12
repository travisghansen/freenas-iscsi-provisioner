package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &Initiator{}
)

type Initiator struct {
	Id          int    `json:"id,omitempty"`
	Tag         int    `json:"iscsi_target_initiator_tag,omitempty"`
	AuthNetwork string `json:"iscsi_target_initiator_auth_network,omitempty"`
	Comment     string `json:"iscsi_target_Initiator_comment,omitempty"`
	Initiators  string `json:"iscsi_target_initiator_initiators,omitempty"`
}

func (i *Initiator) CopyFrom(source FreenasResource) error {
	src, ok := source.(*Initiator)
	if ok {
		i.Id = src.Id
		i.Tag = src.Tag
		i.AuthNetwork = src.AuthNetwork
		i.Comment = src.Comment
		i.Initiators = src.Initiators
	}

	return errors.New("Cannot copy, src is not a Initiator")
}

func (i *Initiator) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authorizedinitiator/%d/", i.Id)
	var initiator Initiator
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&initiator)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	i.CopyFrom(&initiator)

	return nil
}

func (i *Initiator) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/authorizedinitiator/"
	var initiator Initiator
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(i).Receive(&initiator, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating initiator for %+v - %v", *i, body))
	}

	i.CopyFrom(&initiator)

	return nil
}

func (i *Initiator) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authorizedinitiator/%d/", i.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
