package freenas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &Initiator{}
)

// Initiator represents an Initiator instance
type Initiator struct {
	ID          int    `json:"id,omitempty"`
	Tag         int    `json:"iscsi_target_initiator_tag,omitempty"`
	AuthNetwork string `json:"iscsi_target_initiator_auth_network,omitempty"`
	Comment     string `json:"iscsi_target_Initiator_comment,omitempty"`
	Initiators  string `json:"iscsi_target_initiator_initiators,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (i *Initiator) CopyFrom(source Resource) error {
	src, ok := source.(*Initiator)
	if ok {
		i.ID = src.ID
		i.Tag = src.Tag
		i.AuthNetwork = src.AuthNetwork
		i.Comment = src.Comment
		i.Initiators = src.Initiators
	}

	return errors.New("Cannot copy, src is not a Initiator")
}

// Get gets an Initiator instance
func (i *Initiator) Get(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authorizedinitiator/%d/", i.ID)
	var initiator Initiator
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&initiator)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	i.CopyFrom(&initiator)

	return resp, nil
}

// Create creates an Initiator instance
func (i *Initiator) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/authorizedinitiator/"
	var initiator Initiator
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(i).Receive(&initiator, nil)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating initiator for %+v - %v", *i, body)
	}

	i.CopyFrom(&initiator)

	return resp, nil
}

// Delete deletes an Initiator instance
func (i *Initiator) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authorizedinitiator/%d/", i.ID)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return resp, err
}
