package freenas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	_ Resource = &Dataset{}
)

// Dataset represents an zfs dataset
type Dataset struct {
	Avail          int64  `json:"avail,omitempty"`
	Mountpoint     string `json:"mountpoint,omitempty"`
	Name           string `json:"name"`
	Pool           string `json:"pool"`
	Recordsize     int64  `json:"recordsize,omitempty"`
	Refquota       int64  `json:"refquota,omitempty"`
	Refreservation int64  `json:"refreservation,omitempty"`
	Refer          int64  `json:"refer,omitempty"`
	Used           int64  `json:"used,omitempty"`
	Comments       string `json:"comments,omitempty"`
}

func (d *Dataset) String() string {
	return filepath.Join(d.Pool, d.Name)
}

// CopyFrom copies data from a response into an existing resource instance
func (d *Dataset) CopyFrom(source Resource) error {
	src, ok := source.(*Dataset)
	if ok {
		d.Avail = src.Avail
		d.Mountpoint = src.Mountpoint
		d.Name = src.Name
		d.Pool = src.Pool
		d.Recordsize = src.Recordsize
		d.Refquota = src.Refquota
		d.Refreservation = src.Refreservation
		d.Refer = src.Refer
		d.Used = src.Used
		d.Comments = src.Comments
	}

	return errors.New("Cannot copy, src is not a Dataset")
}

// Get gets a Dataset instance
func (d *Dataset) Get(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/storage/dataset/%s/", d.Name)
	var dataset Dataset
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&dataset)
	if err != nil {
		glog.Warningln(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error getting dataset \"%s\" - message: %v, status: %d", d.Name, body, resp.StatusCode)
	}

	d.CopyFrom(&dataset)

	return resp, nil
}

// Create creates a Dataset instance
func (d *Dataset) Create(server *Server) (*http.Response, error) {
	parent, dsName := filepath.Split(d.Name)
	endpoint := fmt.Sprintf("/api/v1.0/storage/dataset/%s", parent)
	var dataset Dataset

	// rewrite Name attribute to support crazy api semantics
	d.Name = dsName

	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(d).Receive(&dataset, nil)

	// rewrite Name attribute to support crazy api semantics
	d.Name = filepath.Join(parent, dsName)

	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("Error creating dataset \"%s\" - message: %v, status: %d", d.Name, body, resp.StatusCode)
	}

	d.CopyFrom(&dataset)

	return resp, nil
}

// Delete deletes a Dataset instance
func (d *Dataset) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/storage/dataset/%s/", d.Name)
	var e string
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, &e)
	if err != nil {
		glog.Warningln(err)
		return nil, err
	}

	if resp.StatusCode != 204 {
		return resp, fmt.Errorf("Error deleting Dataset \"%s\" - message: %s, status: %d", d.Name, e, resp.StatusCode)
	}

	return resp, nil
}
