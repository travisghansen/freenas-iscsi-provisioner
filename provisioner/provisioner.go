package provisioner

import (
	//"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/travisghansen/freenas-iscsi-provisioner/freenas"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// freenasProvisioner is an implem of controller.Provisioner
	_ controller.Provisioner = &freenasProvisioner{}
)

type freenasProvisionerConfig struct {
	// common params
	FSType string

	// Provisioner options
	ProvisionerTargetPortal             string
	ProvisionerPortals                  string
	ProvisionerEnableDeterministicNames bool
	ProvisionerISCSINamePrefix          string
	ProvisionerISCSINameSuffix          string
	ProvisionerISCSIInterface           string

	// Dataset options
	DatasetParentName       string
	DatasetEnableNamespaces bool

	// TargetGroup options
	TargetGroupAuthgroup      int
	TargetGroupAuthtype       string
	TargetGroupInitiatorgroup int
	TargetGroupPortalgroup    int

	// Zvol options
	ZvolCompression string
	ZvolDedup       string
	ZvolSparse      bool
	ZvolForce       bool
	ZvolBlocksize   string

	// Extent options
	ExtentBlocksize                int
	ExtentDisablePhysicalBlocksize bool
	ExtentAvailThreshold           int
	ExtentInsecureTpc              bool
	ExtentXen                      bool
	ExtentRpm                      string
	ExtentReadOnly                 bool

	// Server options
	ServerSecretNamespace string
	ServerSecretName      string
	ServerProtocol        string
	ServerHost            string
	ServerPort            int
	ServerUsername        string
	ServerPassword        string
	ServerAllowInsecure   bool
}

func (p *freenasProvisioner) GetConfig(storageClassName string) (*freenasProvisionerConfig, error) {
	class, err := p.Client.StorageV1beta1().StorageClasses().Get(storageClassName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var fsType string = "ext4"

	// provisioner defaults
	var provisionerTargetPortal string = ""
	var provisionerPortals string = ""
	var provisionerEnableDeterministicNames bool = true
	var provisionerISCSINamePrefix string = ""
	var provisionerISCSINameSuffix string = ""
	var provisionerISCSIInterface string = "default"

	// dataset defaults
	var datasetParentName string = "tank"
	var datasetEnableNamespaces bool = true

	// targetGroup defaults
	var targetGroupAuthgroup int
	var targetGroupAuthtype string = "None"
	var targetGroupInitiatorgroup int
	var targetGroupPortalgroup int

	// zvol defaults
	var zvolCompression string = ""
	var zvolDedup string = ""
	var zvolSparse bool = true
	var zvolForce bool = false
	var zvolBlocksize string = ""

	// extent defaults
	var extentBlocksize int = 0
	var extentDisablePhysicalBlocksize bool = true
	var extentAvailThreshold int = 0
	var extentInsecureTpc bool = true
	var extentXen bool = false
	var extentRpm string = ""
	var extentReadOnly bool = false

	// server options
	var serverSecretNamespace string = "kube-system"
	var serverSecretName string = "freenas-iscsi"
	var serverProtocol string = "http"
	var serverHost string = "localhost"
	var serverPort int = 80
	var serverUsername string = "root"
	var serverPassword string = ""
	var serverAllowInsecure bool = false

	// set values from StorageClass parameters
	for k, v := range class.Parameters {
		switch k {
		case "fsType":
			fsType = v

		// Provisioner options
		case "provisionerTargetPortal":
			provisionerTargetPortal = v
		case "provisionerPortals":
			provisionerPortals = v
		case "provisionerEnableDeterministicNames":
			provisionerEnableDeterministicNames, _ = strconv.ParseBool(v)
		case "provisionerISCSINamePrefix":
			provisionerISCSINamePrefix = v
		case "provisionerISCSINameSuffix":
			provisionerISCSINameSuffix = v
		case "provisionerISCSIInterface":
			provisionerISCSIInterface = v

		// Dataset options
		case "datasetParentName":
			datasetParentName = v
		case "datasetEnableNamespaces":
			datasetEnableNamespaces, _ = strconv.ParseBool(v)

		// TargetGroup options
		case "targetGroupAuthgroup":
			targetGroupAuthgroup, _ = strconv.Atoi(v)
		case "targetGroupAuthtype":
			targetGroupAuthtype = v
		case "targetGroupInitiatorgroup":
			targetGroupInitiatorgroup, _ = strconv.Atoi(v)
		case "targetGroupPortalgroup":
			targetGroupPortalgroup, _ = strconv.Atoi(v)

		// Zvol options
		case "zvolCompression":
			zvolCompression = v
		case "zvolDedup":
			zvolDedup = v
		case "zvolSparse":
			zvolSparse, _ = strconv.ParseBool(v)
		case "zvolForce":
			zvolForce, _ = strconv.ParseBool(v)
		case "zvolBlocksize":
			zvolBlocksize = v

		// Extent options
		case "extentBlocksize":
			extentBlocksize, _ = strconv.Atoi(v)
		case "extentDisablePhysicalBlocksize":
			extentDisablePhysicalBlocksize, _ = strconv.ParseBool(v)
		case "extentAvailThreshold":
			extentAvailThreshold, _ = strconv.Atoi(v)
		case "extentInsecureTpc":
			extentInsecureTpc, _ = strconv.ParseBool(v)
		case "extentXen":
			extentXen, _ = strconv.ParseBool(v)
		case "extentRpm":
			extentRpm = v
		case "extentReadOnly":
			extentReadOnly, _ = strconv.ParseBool(v)

		// Server options
		case "serverSecretNamespace":
			serverSecretNamespace = v
		case "serverSecretName":
			serverSecretName = v
		}
	}

	secret, err := p.GetSecret(serverSecretNamespace, serverSecretName)
	if err != nil {
		return nil, err
	}

	// set values from secret
	for k, v := range secret.Data {
		switch k {
		case "protocol":
			serverProtocol = BytesToString(v)
		case "host":
			serverHost = BytesToString(v)
		case "port":
			serverPort, _ = strconv.Atoi(BytesToString(v))
		case "username":
			serverUsername = BytesToString(v)
		case "password":
			serverPassword = BytesToString(v)
		case "allowInsecure":
			serverAllowInsecure, _ = strconv.ParseBool(BytesToString(v))
		}
	}

	if provisionerTargetPortal == "" {
		provisionerTargetPortal = serverHost + ":3260"
	}

	return &freenasProvisionerConfig{
		FSType: fsType,

		// Provisioner options
		ProvisionerTargetPortal:             provisionerTargetPortal,
		ProvisionerPortals:                  provisionerPortals,
		ProvisionerEnableDeterministicNames: provisionerEnableDeterministicNames,
		ProvisionerISCSINamePrefix:          provisionerISCSINamePrefix,
		ProvisionerISCSINameSuffix:          provisionerISCSINameSuffix,
		ProvisionerISCSIInterface:           provisionerISCSIInterface,

		// Dataset options
		DatasetParentName:       datasetParentName,
		DatasetEnableNamespaces: datasetEnableNamespaces,

		// TargetGroup options
		TargetGroupAuthgroup:      targetGroupAuthgroup,
		TargetGroupAuthtype:       targetGroupAuthtype,
		TargetGroupInitiatorgroup: targetGroupInitiatorgroup,
		TargetGroupPortalgroup:    targetGroupPortalgroup,

		// Zvol options
		ZvolCompression: zvolCompression,
		ZvolDedup:       zvolDedup,
		ZvolSparse:      zvolSparse,
		ZvolForce:       zvolForce,
		ZvolBlocksize:   zvolBlocksize,

		// Extent options
		ExtentBlocksize:                extentBlocksize,
		ExtentDisablePhysicalBlocksize: extentDisablePhysicalBlocksize,
		ExtentAvailThreshold:           extentAvailThreshold,
		ExtentInsecureTpc:              extentInsecureTpc,
		ExtentXen:                      extentXen,
		ExtentRpm:                      extentRpm,
		ExtentReadOnly:                 extentReadOnly,

		// Server options
		ServerSecretNamespace: serverSecretNamespace,
		ServerSecretName:      serverSecretName,
		ServerProtocol:        serverProtocol,
		ServerHost:            serverHost,
		ServerPort:            serverPort,
		ServerUsername:        serverUsername,
		ServerPassword:        serverPassword,
		ServerAllowInsecure:   serverAllowInsecure,
	}, nil
}

type freenasProvisioner struct {
	Client     kubernetes.Interface
	Identifier string
}

func New(client kubernetes.Interface, identifier string) controller.Provisioner {
	return &freenasProvisioner{
		Client:     client,
		Identifier: identifier,
	}
}

func AccessModesContains(modes []v1.PersistentVolumeAccessMode, mode v1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

func AccessModesContainedInAll(indexedModes []v1.PersistentVolumeAccessMode, requestedModes []v1.PersistentVolumeAccessMode) bool {
	for _, mode := range requestedModes {
		if !AccessModesContains(indexedModes, mode) {
			return false
		}
	}
	return true
}

func (p *freenasProvisioner) getAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
		v1.ReadOnlyMany,
	}
}

func (p *freenasProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	if !AccessModesContainedInAll(p.getAccessModes(), options.PVC.Spec.AccessModes) {
		return nil, fmt.Errorf("invalid AccessModes %v: only AccessModes %v are supported", options.PVC.Spec.AccessModes, p.getAccessModes())
	}

	var err error

	// get config
	config, err := p.GetConfig(*options.PVC.Spec.StorageClassName)
	if err != nil {
		return nil, err
	}
	//glog.Infof("%+v\n", config)

	// get server
	freenasServer, err := p.GetServer(*config)
	if err != nil {
		return nil, err
	}

	// get iscsi configuration
	iscsiConfig := freenas.ISCSIConfig{}
	err = iscsiConfig.Get(freenasServer)
	if err != nil {
		return nil, err
	}

	// get parent dataset
	parentDs := freenas.Dataset{
		Name: config.DatasetParentName,
	}
	err = parentDs.Get(freenasServer)
	if err != nil {
		return nil, err
	}

	meta := options.PVC.GetObjectMeta()
	zvolName := options.PVName
	iscsiName := options.PVName
	dsNamespace := ""

	if config.DatasetEnableNamespaces {
		dsNamespace = meta.GetNamespace()
	}

	if config.ProvisionerEnableDeterministicNames {
		iscsiName = meta.GetNamespace() + "-" + meta.GetName()
		if config.DatasetEnableNamespaces {
			zvolName = meta.GetNamespace() + "/" + meta.GetName()
		} else {
			zvolName = meta.GetNamespace() + "-" + meta.GetName()
		}
	}

	zvolName = strings.TrimPrefix(parentDs.Name, parentDs.Pool+"/") + "/" + zvolName
	iscsiName = config.ProvisionerISCSINamePrefix + iscsiName + config.ProvisionerISCSINameSuffix

	glog.Infof("Creating target: \"%s\", zvol: \"%s/%s\", extent: \"%s\", ", iscsiName, parentDs.Pool, zvolName, iscsiName)

	// Create namepsace dataset if desired
	if config.DatasetEnableNamespaces {
		nsDs := freenas.Dataset{
			Pool:     parentDs.Pool,
			Name:     filepath.Join(parentDs.Name, dsNamespace),
			Comments: "k8s provisioned namespace",
		}

		err = nsDs.Get(freenasServer)
		if err != nil {
			glog.Infof("creating namespace dataset \"%s\"", nsDs.Name)
			err = nsDs.Create(freenasServer)
		} else {
			glog.Infof("namespace dataset \"%s\" already exists", nsDs.Name)
		}
	}
	if err != nil {
		return nil, err
	}

	// Create target
	target := freenas.Target{
		Name:  iscsiName,
		Alias: "",
		Mode:  "iscsi",
	}
	err = target.Create(freenasServer)
	if err != nil {
		return nil, err
	}

	// Create targetgroup(s)
	targetGroup := freenas.TargetGroup{
		Target:         target.Id,
		Authgroup:      config.TargetGroupAuthgroup,
		Authtype:       config.TargetGroupAuthtype,
		Initialdigest:  "Auto",
		Initiatorgroup: config.TargetGroupInitiatorgroup,
		Portalgroup:    config.TargetGroupPortalgroup,
	}
	err = targetGroup.Create(freenasServer)
	if err != nil {
		target.Delete(freenasServer)
		return nil, err
	}

	// Create zvol
	var zvolVolsize int64 = 0
	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	zvolVolsize = volSize.Value()
	zvolVolsizeGB := (float64(zvolVolsize) / 1024 / 1024 / 1024)

	zvol := freenas.Zvol{
		Name:        zvolName,
		Comments:    "comments",
		Compression: config.ZvolCompression, // config - "" (inherit), lz4, gzip-9, etc
		Dedup:       config.ZvolDedup,       // config - nil (inherit), on, off, verify
		Volsize:     strconv.FormatFloat(zvolVolsizeGB, 'f', -1, 64) + " GiB",
		Sparse:      config.ZvolSparse,
		Force:       config.ZvolForce,
		Blocksize:   config.ZvolBlocksize, // config - 512, 1K, 2K, 4K, 8K, 16K, 32K, 64K, 128K
		Dataset:     parentDs,
	}
	err = zvol.Create(freenasServer)
	if err != nil {
		target.Delete(freenasServer)
		return nil, err
	}

	// Create extent
	extent := freenas.Extent{
		Name:           iscsiName,
		Type:           "Disk",
		Disk:           "zvol/" + parentDs.Pool + "/" + zvol.Name,
		Blocksize:      config.ExtentBlocksize, //config - 512, 1024, 2048, or 4096
		Pblocksize:     config.ExtentDisablePhysicalBlocksize,
		AvailThreshold: config.ExtentAvailThreshold,
		Comment:        "some comments",
		InsecureTpc:    config.ExtentInsecureTpc,
		Xen:            config.ExtentXen,
		Rpm:            config.ExtentRpm, // config - Unknown, SSD, 5400, 7200, 10000, 15000
		Ro:             config.ExtentReadOnly,
	}
	err = extent.Create(freenasServer)
	if err != nil {
		zvol.Delete(freenasServer)
		target.Delete(freenasServer)
		return nil, err
	}

	// Create targettoextent
	lunid := 0
	targetToExtent := freenas.TargetToExtent{
		Extent: extent.Id,
		Lunid:  &lunid,
		Target: target.Id,
	}
	err = targetToExtent.Create(freenasServer)
	if err != nil {
		extent.Delete(freenasServer)
		zvol.Delete(freenasServer)
		target.Delete(freenasServer)
		return nil, err
	}

	var portals []string
	if len(config.ProvisionerPortals) > 0 {
		portals = strings.Split(config.ProvisionerPortals, ",")
	}

	var lun int32 = 0

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"freenasISCSIProvisionerIdentity": p.Identifier,
				"datasetParent":                   config.DatasetParentName,
				"pool":                            parentDs.Pool,
				"zvol":                            zvolName,
				"iscsiName":                       iscsiName,
				"targetId":                        strconv.Itoa(target.Id),
				"extentId":                        strconv.Itoa(extent.Id),
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			// set volumeMode from PVC Spec
			//VolumeMode: options.PVC.Spec.VolumeMode,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIVolumeSource{
					TargetPortal:   config.ProvisionerTargetPortal,
					Portals:        portals,
					IQN:            iscsiConfig.Basename + ":" + iscsiName,
					ISCSIInterface: config.ProvisionerISCSIInterface,
					Lun:            lun,
					ReadOnly:       extent.Ro,
					FSType:         config.FSType,
					//DiscoveryCHAPAuth: false,
					//SessionCHAPAuth:   false,
					//SecretRef:         getSecretRef(getBool(options.Parameters["chapAuthDiscovery"]), getBool(options.Parameters["chapAuthSession"]), &v1.SecretReference{Name: viper.GetString("provisioner-name") + "-chap-secret"}),
				},
			},
		},
	}

	return pv, nil
}

func (p *freenasProvisioner) Delete(volume *v1.PersistentVolume) error {
	var targetId, extentId int
	var poolName, zvolName, iscsiName, datasetParentName string

	targetIdAnnotation, ok := volume.Annotations["targetId"]
	if ok {
		targetId, _ = strconv.Atoi(targetIdAnnotation)
	}

	extentIdAnnotation, ok := volume.Annotations["extentId"]
	if ok {
		extentId, _ = strconv.Atoi(extentIdAnnotation)
	}

	poolName = volume.Annotations["pool"]
	zvolName = volume.Annotations["zvol"]
	iscsiName = volume.Annotations["iscsiName"]
	datasetParentName = volume.Annotations["datasetParent"]

	var err error

	// get config
	config, err := p.GetConfig(volume.Spec.StorageClassName)
	if err != nil {
		return err
	}
	//glog.Infof("%+v\n", config)

	// get server
	freenasServer, err := p.GetServer(*config)
	if err != nil {
		return err
	}

	// get parent dataset
	parentDs := freenas.Dataset{
		Name: datasetParentName,
	}
	err = parentDs.Get(freenasServer)
	if err != nil {
		return err
	}

	glog.Infof("Deleting target: %d (\"%s\"), extent: %d (\"%s\"), zvol: \"%s/%s\"", targetId, iscsiName, extentId, iscsiName, poolName, zvolName)

	// Delete target
	// NOTE: deletting a target inherently deletes associated targetgroup(s) and targettoextent(s)
	target := freenas.Target{
		Id: targetId,
	}
	target.Delete(freenasServer)

	// Delete extent
	extent := freenas.Extent{
		Id: extentId,
	}
	extent.Delete(freenasServer)

	// Delete zvol
	zvol := freenas.Zvol{
		Name:    zvolName,
		Dataset: parentDs,
	}
	zvol.Delete(freenasServer)

	return nil
}

func (p *freenasProvisioner) GetServer(config freenasProvisionerConfig) (*freenas.FreenasServer, error) {
	return freenas.NewFreenasServer(
		config.ServerProtocol, config.ServerHost, config.ServerPort,
		config.ServerUsername, config.ServerPassword,
		config.ServerAllowInsecure,
	), nil
}

func (p *freenasProvisioner) GetSecret(namespace, secretName string) (*v1.Secret, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("Cannot get kube client")
	}
	return p.Client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
}

func BytesToString(data []byte) string {
	return string(data[:])
}

// Prep for resizing
// https://github.com/kubernetes-incubator/external-storage/blob/master/gluster/file/cmd/glusterfile-provisioner/glusterfile-provisioner.go#L433
/*
func (p *freenasProvisioner) RequiresFSResize() bool {
	return false
}

func (p *glusterfileProvisioner) ExpandVolumeDevice(spec *volume.Spec, newSize resource.Quantity, oldSize resource.Quantity) (resource.Quantity, error) {
	return newVolumeSize, nil
}

func (p *iscsiProvisioner) SupportsBlock() bool {
	return true
}
*/
