package freenas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &AuthCredential{}
)

// AuthCredential represents an ISCSI credential
type AuthCredential struct {
	ID         int    `json:"id,omitempty"`
	Tag        int    `json:"tag,omitempty"`
	User       string `json:"iscsi_target_auth_user,omitempty"`
	Secret     string `json:"iscsi_target_auth_secret,omitempty"`
	Peeruser   string `json:"iscsi_target_auth_peeruser,omitempty"`
	Peersecret string `json:"iscsi_target_auth_peersecret,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (a *AuthCredential) CopyFrom(source Resource) error {
	src, ok := source.(*AuthCredential)
	if ok {
		a.ID = src.ID
		a.Tag = src.Tag
		a.User = src.User
		a.Secret = src.Secret
		a.Peeruser = src.Peeruser
		a.Peersecret = src.Peersecret
	}

	return errors.New("Cannot copy, src is not a AuthCredential")
}

// Get gets an AuthCredential instance
func (a *AuthCredential) Get(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authcredential/%d/", a.ID)
	var authCredential AuthCredential
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&authCredential)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	a.CopyFrom(&authCredential)

	return resp, nil
}

// Create creates an AuthCredential instance
func (a *AuthCredential) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/authcredential/"
	var authCredential AuthCredential
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(a).Receive(&authCredential, nil)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating authcredential for %+v - %v", *a, body)
	}

	a.CopyFrom(&authCredential)

	return resp, nil
}

// Delete deletes an AuthCredential instance
func (a *AuthCredential) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/authcredential/%d/", a.ID)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return resp, err
}
