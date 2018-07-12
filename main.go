package main

import (
	"github.com/travisghansen/freenas-iscsi-provisioner/cli"
)

const (
	AppName = "freenas-iscsi-provisioner"
	AppDesc = "Kubernetes FreeNAS Provisioner (iscsi)"
)

var (
	AppVersion string
)

func main() {
	if AppVersion == "" {
		AppVersion = "master"
	}

	cli.Process(AppName, AppDesc, AppVersion)
}
