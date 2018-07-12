package freenas

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

var (
	_ FreenasResource = &Extent{}
)

type Extent struct {
	Id             int    `json:"id,omitempty"`
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

func (e *Extent) CopyFrom(source FreenasResource) error {
	src, ok := source.(*Extent)
	if ok {
		e.Id = src.Id
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

func (e *Extent) Get(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/extent/%d/", e.Id)
	var extent Extent
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&extent)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	e.CopyFrom(&extent)

	return nil
}

func (e *Extent) Create(server *FreenasServer) error {
	endpoint := "/api/v1.0/services/iscsi/extent/"
	var extent Extent
	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(e).Receive(&extent, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating extent for %+v - %v", *e, body))
	}

	e.CopyFrom(&extent)

	return nil
}

func (e *Extent) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/services/iscsi/extent/%d/", e.Id)
	_, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
	}
	return err
}
