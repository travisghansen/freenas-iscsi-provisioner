package freenas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &TargetToExtent{}
)

// TargetToExtent represents a TargetToExtent instance
type TargetToExtent struct {
	ID     int  `json:"id,omitempty"`
	Extent int  `json:"iscsi_extent,omitempty"`
	Lunid  *int `json:"iscsi_lunid,omitempty"`
	Target int  `json:"iscsi_target,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (t *TargetToExtent) CopyFrom(source Resource) error {
	src, ok := source.(*TargetToExtent)
	if ok {
		t.ID = src.ID
		t.Extent = src.Extent
		t.Lunid = src.Lunid
		t.Target = src.Target
	}

	return errors.New("Cannot copy, src is not a TargetToExtent")
}

// Get gets a TargetToExtent instance
func (t *TargetToExtent) Get(server *Server) (*http.Response, error) {
	if t.ID > 0 {
		endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targettoextent/%d/", t.ID)
		var targetToExtent TargetToExtent
		resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&targetToExtent)
		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		t.CopyFrom(&targetToExtent)

		return resp, nil
	}

	// find target/extent/lun ID
	if t.Extent > 0 && t.Target > 0 && *t.Lunid >= 0 {
		endpoint := "/api/v1.0/services/iscsi/targettoextent/?limit=1000"
		var list []TargetToExtent
		var e interface{}
		resp, err := server.getSlingConnection().Get(endpoint).Receive(&list, &e)

		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		if resp.StatusCode != 200 {
			body, _ := json.Marshal(e)
			return resp, fmt.Errorf("Error getting TargetToExtent extent ID: %d, lunid ID: %d, target ID: %d - message: %v, status: %d", t.Extent, *t.Lunid, t.Target, string(body), resp.StatusCode)
		}

		for _, item := range list {
			if item.Target == t.Target && item.Extent == t.Extent && *item.Lunid == *t.Lunid {
				t.CopyFrom(&item)
				glog.Infof("found TargetToExtent extent ID: %d, lunid ID: %d, target ID: %d - %+v", t.Extent, *t.Lunid, t.Target, *t)
				return resp, nil
			}
		}
	}

	// Nothing found
	return nil, errors.New("no TargetToExtent has been found")
}

// Create creates a TargetToExtent instance
func (t *TargetToExtent) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/targettoextent/"
	var targetToExtent TargetToExtent
	var e interface{}
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(t).Receive(&targetToExtent, &e)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := json.Marshal(e)
		return resp, fmt.Errorf("Error creating TargetToExtent for %+v - message: %s, status: %d", *t, string(body), resp.StatusCode)
	}

	t.CopyFrom(&targetToExtent)

	return resp, nil
}

// Delete deletes a TargetToExtent instance
func (t *TargetToExtent) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/targettoextent/%d/", t.ID)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return resp, err
}
