package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &AuthCredential{}
)

type AuthCredential struct {
	Id         int    `json:"id,omitempty"`
	Tag        int    `json:"tag,omitempty"`
	User       string `json:"iscsi_target_auth_user,omitempty"`
	Secret     string `json:"iscsi_target_auth_secret,omitempty"`
	Peeruser   string `json:"iscsi_target_auth_peeruser,omitempty"`
	Peersecret string `json:"iscsi_target_auth_peersecret,omitempty"`
}

func (a *AuthCredential) CopyFrom(source FreenasResource) error {
	src, ok := source.(*AuthCredential)
	if ok {
		a.Id = src.Id
		a.Tag = src.Tag
		a.User = src.User
		a.Secret = src.Secret
		a.Peeruser = src.Peeruser
		a.Peersecret = src.Peersecret
	}

	return errors.New("Cannot copy, src is not a AuthCredential")
}

func (a *AuthCredential) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authcredential/%d/", a.Id)
	var authCredential AuthCredential
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&authCredential)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	a.CopyFrom(&authCredential)

	return nil
}

func (a *AuthCredential) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/authcredential/"
	var authCredential AuthCredential
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(a).Receive(&authCredential, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating authcredential for %+v - %v", *a, body))
	}

	a.CopyFrom(&authCredential)

	return nil
}

func (a *AuthCredential) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authcredential/%d/", a.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
