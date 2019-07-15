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
	_ Resource = &Extent{}
)

// Extent represents an ISCSI extent
type Extent struct {
	ID             int    `json:"id,omitempty"`
	AvailThreshold int    `json:"iscsi_target_extent_avail_threshold,omitempty"`
	Blocksize      int    `json:"iscsi_target_extent_blocksize,omitempty"`
	Comment        string `json:"iscsi_target_extent_comment,omitempty"`
	Filesize       string `json:"iscsi_target_extent_filesize,omitempty"`
	InsecureTpc    bool   `json:"iscsi_target_extent_insecure_tpc,omitempty"`
	Legacy         bool   `json:"iscsi_target_extent_legacy,omitempty"`
	Naa            string `json:"iscsi_target_extent_naa,omitempty"`
	Name           string `json:"iscsi_target_extent_name,omitempty"`
	Path           string `json:"iscsi_target_extent_path,omitempty"`
	Disk           string `json:"iscsi_target_extent_disk,omitempty"`
	Pblocksize     bool   `json:"iscsi_target_extent_pblocksize,omitempty"`
	Ro             bool   `json:"iscsi_target_extent_ro,omitempty"`
	Rpm            string `json:"iscsi_target_extent_rpm,omitempty"`
	Serial         string `json:"iscsi_target_extent_serial,omitempty"`
	Type           string `json:"iscsi_target_extent_type,omitempty"`
	Xen            bool   `json:"iscsi_target_extent_xen,omitempty"`
}

// CopyFrom copies data from a response into an existing resource instance
func (e *Extent) CopyFrom(source Resource) error {
	src, ok := source.(*Extent)
	if ok {
		e.ID = src.ID
		e.AvailThreshold = src.AvailThreshold
		e.Blocksize = src.Blocksize
		e.Comment = src.Comment
		e.Filesize = src.Filesize
		e.InsecureTpc = src.InsecureTpc
		e.Legacy = src.Legacy
		e.Naa = src.Naa
		e.Name = src.Name
		e.Path = src.Path
		e.Pblocksize = src.Pblocksize
		e.Ro = src.Ro
		e.Rpm = src.Rpm
		e.Serial = src.Serial
		e.Type = src.Type
		e.Xen = src.Xen
	}

	return errors.New("Cannot copy, src is not a Extent")
}

// Get gets an Extent instance
func (e *Extent) Get(server *Server) (*http.Response, error) {
	if e.ID > 0 {
		endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/extent/%d/", e.ID)
		var extent Extent
		resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&extent)
		if err != nil {
			glog.Warningln(err)
			return resp, err
		}

		e.CopyFrom(&extent)

		return resp, nil
	}

	// find by name
	if len(e.Name) > 0 {
		endpoint := "/api/v1.0/services/iscsi/extent/?limit=1000"
		var list []Extent
		var es interface{}
		resp, err := server.getSlingConnection().Get(endpoint).Receive(&list, &es)

		if err != nil {
			glog.Warningln(err)
			return nil, err
		}

		if resp.StatusCode != 200 {
			body, _ := json.Marshal(es)
			return resp, fmt.Errorf("Error getting Extent \"%s\" - message: %v, status: %d", e.Name, string(body), resp.StatusCode)
		}

		for _, item := range list {
			if item.Name == e.Name {
				e.CopyFrom(&item)
				glog.Infof("found Extent name: %s - %+v", e.Name, *e)
				return resp, nil
			}
		}
	}

	// Nothing found
	return nil, errors.New("no Extent has been found")
}

// Create creates an Extent instance
func (e *Extent) Create(server *Server) (*http.Response, error) {
	endpoint := "/api/v1.0/services/iscsi/extent/"
	var extent Extent
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(e).Receive(&extent, nil)
	if err != nil {
		glog.Warningln(err)
		return nil, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating extent for %+v - %v", *e, body)
	}

	e.CopyFrom(&extent)

	return resp, nil
}

// Delete deletes an Extent instance
func (e *Extent) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/extent/%d/", e.ID)
	var es string
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, &es)
	if err != nil {
		glog.Warningln(err)
	}

	if resp.StatusCode != 204 {
		return resp, fmt.Errorf("Error deleting Extent - message: %s, status: %d", es, resp.StatusCode)
	}

	return resp, nil
}
