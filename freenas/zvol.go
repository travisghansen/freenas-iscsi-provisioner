package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &Zvol{}
)

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

func (z *Zvol) CopyFrom(source FreenasResource) error {
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

func (z *Zvol) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/%s/", z.Dataset.Pool, z.Name)
	var zvol Zvol
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&zvol)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	z.CopyFrom(&zvol)

	return nil
}

func (z *Zvol) Create(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/", z.Dataset.Pool)
	//var zvol Zvol
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(z).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating zvol for %+v - %v", *z, body))
	}

	//z.CopyFrom(&zvol)

	return nil
}

func (z *Zvol) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/zvols/%s/", z.Dataset.Pool, z.Name)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
