package freenas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &Target{}
)

// Target represents a Target instance
type Target struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"iscsi_target_name,omitempty"`
	Alias string `json:"iscsi_target_alias,omitempty"`
	Mode  string `json:"iscsi_target_mode,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (t *Target) CopyFrom(source Resource) error {
	src, ok := source.(*Target)
	if ok {
		t.ID = src.ID
		t.Name = src.Name
		t.Alias = src.Alias
		t.Mode = src.Mode
	}

	return errors.New("Cannot copy, src is not a Target")
}

// Get gets a Target instance
func (t *Target) Get(server *Server) (*http.Response, error) {
	// find by ID
	if t.ID > 0 {
		endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/target/%d/", t.ID)
		var target Target
		resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&target)
		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		t.CopyFrom(&target)

		return resp, nil
	}

	// find by name
	if len(t.Name) > 0 {
		endpoint := "/api/v1.0/services/iscsi/target/?limit=1000"
		var list []Target
		var e interface{}
		resp, err := server.getSlingConnection().Get(endpoint).Receive(&list, &e)

		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		if resp.StatusCode != 200 {
			body, _ := json.Marshal(e)
			return resp, fmt.Errorf("Error getting target \"%s\" - message: %v, status: %d", t.Name, string(body), resp.StatusCode)
		}

		for _, item := range list {
			if item.Name == t.Name {
				t.CopyFrom(&item)
				glog.Infof("found Target name: %s - %+v", t.Name, *t)
				return resp, nil
			}
		}
	}

	// Nothing found
	return nil, errors.New("no Target has been found")
}

// GetByName gets a Target instance
func (t *Target) GetByName(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/target/%d/", t.ID)
	var target Target
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&target)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	t.CopyFrom(&target)

	return nil, nil
}

// Create creates a Target instance
func (t *Target) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/target/"
	var target Target
	var e interface{}
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&target, &e)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := json.Marshal(e)
		return resp, fmt.Errorf("Error creating Target for %+v - message: %s, status: %d", *t, string(body), resp.StatusCode)
	}

	t.CopyFrom(&target)

	return resp, nil
}

// Delete deletes a Target instance
func (t *Target) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/target/%d/", t.ID)
	var e string
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, &e)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 204 {
		return resp, fmt.Errorf("Error deleting Target - message: %s, status: %d", e, resp.StatusCode)
	}

	return resp, nil
}
