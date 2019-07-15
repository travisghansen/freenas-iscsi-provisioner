package freenas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

var (
	_ Resource = &Zvol{}
)

// Zvol represents a zvol instance
type Zvol struct {
	Name        string `json:"name,omitempty"`
	Avail       int64  `json:"avail,omitempty"`
	Comments    string `json:"comments,omitempty"`
	Compression string `json:"compression,omitempty"`
	Dedup       string `json:"dedup,omitempty"`
	Refer       int64  `json:"refer,omitempty"`
	Used        int64  `json:"used,omitempty"`
	//Volsize     int64  `json:"volsize,omitempty"`
	Volsize   string  `json:"volsize,omitempty,"`
	Sparse    bool    `json:"sparse,omitempty"`
	Force     bool    `json:"force,omitempty"`
	Blocksize string  `json:"blocksize,omitempty"`
	Dataset   Dataset `json:"-"`
}

// CopyFrom copies data from a response into an existing resource instance
func (z *Zvol) CopyFrom(source Resource) error {
	src, ok := source.(*Zvol)
	if ok {
		z.Name = src.Name
		z.Avail = src.Avail
		z.Comments = src.Comments
		z.Compression = src.Compression
		z.Dedup = src.Dedup
		z.Refer = src.Refer
		z.Used = src.Used
		z.Volsize = src.Volsize
		z.Sparse = src.Sparse
		z.Force = src.Force
		z.Blocksize = src.Blocksize
	}

	return errors.New("Cannot copy, src is not a Zvol")
}

// Get gets a Zvol instance
func (z *Zvol) Get(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/%s/", z.Dataset.Pool, z.Name)
	var zvol Zvol
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&zvol)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	z.CopyFrom(&zvol)

	return resp, nil
}

// Create creates a Zvol instance
func (z *Zvol) Create(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/", z.Dataset.Pool)
	//var zvol Zvol

	var e interface{}
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(z).Receive(nil, &e)
	if err != nil {
		glog.Warningln(err)
		return resp, err
	}

	if resp.StatusCode != 202 {
		body, _ := json.Marshal(e)
		return resp, fmt.Errorf("Error creating zvol for %+v - message: %s, status: %d", *z, string(body), resp.StatusCode)
	}

	//z.CopyFrom(&zvol)

	return resp, nil
}

// Delete deletes a Zvol instance
func (z *Zvol) Delete(server *Server) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/%s/", z.Dataset.Pool, z.Name)

	var e string
	type DeleteBody struct {
		Cascade bool `json:"cascade"`
	}
	var b = new(DeleteBody)
	b.Cascade = true
	resp, err := server.getSlingConnection().Delete(endpoint).BodyJSON(b).Receive(nil, &e)
	if err != nil {
		glog.Warningln(err)
	}

	if resp.StatusCode != 204 {
		return resp, fmt.Errorf("Error deleting Zvol: %d %s", resp.StatusCode, e)
	}

	return resp, nil
}
