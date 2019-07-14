package provisioner

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/travisghansen/freenas-iscsi-provisioner/freenas"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// freenasProvisioner is an implem of controller.Provisioner
	_ controller.Provisioner = &freenasProvisioner{}
)

type freenasProvisionerConfig struct {
	// common params
	FSType string

	// Provisioner options
	ProvisionerTargetPortal    string
	ProvisionerPortals         string
	ProvisionerISCSINamePrefix string
	ProvisionerISCSINameSuffix string
	ProvisionerISCSIInterface  string

	// Dataset options
	DatasetParentName string

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

	var fsType = "ext4"

	// provisioner defaults
	var provisionerTargetPortal string
	var provisionerPortals string
	var provisionerISCSINamePrefix string
	var provisionerISCSINameSuffix string
	var provisionerISCSIInterface = "default"

	// dataset defaults
	var datasetParentName = "tank"

	// targetGroup defaults
	var targetGroupAuthgroup int
	var targetGroupAuthtype = "None"
	var targetGroupInitiatorgroup int
	var targetGroupPortalgroup int

	// zvol defaults
	var zvolCompression string
	var zvolDedup string
	var zvolSparse = true
	var zvolForce = false
	var zvolBlocksize string

	// extent defaults
	var extentBlocksize int
	var extentDisablePhysicalBlocksize = true
	var extentAvailThreshold int
	var extentInsecureTpc = true
	var extentXen = false
	var extentRpm string
	var extentReadOnly = false

	// server options
	var serverSecretNamespace = "kube-system"
	var serverSecretName = "freenas-iscsi"
	var serverProtocol = "http"
	var serverHost = "localhost"
	var serverPort = 80
	var serverUsername = "root"
	var serverPassword string
	var serverAllowInsecure = false

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
		case "provisionerISCSINamePrefix":
			provisionerISCSINamePrefix = v
		case "provisionerISCSINameSuffix":
			provisionerISCSINameSuffix = v
		case "provisionerISCSIInterface":
			provisionerISCSIInterface = v

		// Dataset options
		case "datasetParentName":
			datasetParentName = v

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
		ProvisionerTargetPortal:    provisionerTargetPortal,
		ProvisionerPortals:         provisionerPortals,
		ProvisionerISCSINamePrefix: provisionerISCSINamePrefix,
		ProvisionerISCSINameSuffix: provisionerISCSINameSuffix,
		ProvisionerISCSIInterface:  provisionerISCSIInterface,

		// Dataset options
		DatasetParentName: datasetParentName,

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

// New creates a new client instance
func New(client kubernetes.Interface, identifier string) controller.Provisioner {
	return &freenasProvisioner{
		Client:     client,
		Identifier: identifier,
	}
}

// AccessModesContains checks if mode is available
func AccessModesContains(modes []v1.PersistentVolumeAccessMode, mode v1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// AccessModesContainedInAll ensures all access modes are available
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
	var resp *http.Response

	var loopErr error
	var loopResp *http.Response

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
	resp, err = iscsiConfig.Get(freenasServer)
	if err != nil {
		return nil, err
	}

	// get parent dataset
	parentDs := freenas.Dataset{
		Name: config.DatasetParentName,
	}
	resp, err = parentDs.Get(freenasServer)
	if err != nil {
		return nil, err
	}

	meta := options.PVC.GetObjectMeta()
	pvcNamespace := meta.GetNamespace()
	pvcName := meta.GetName()
	zvolName := options.PVName
	iscsiName := options.PVName

	zvolName = strings.TrimPrefix(parentDs.Name, parentDs.Pool+"/") + "/" + zvolName
	iscsiName = config.ProvisionerISCSINamePrefix + iscsiName + config.ProvisionerISCSINameSuffix
	extentDiskName := "zvol/" + parentDs.Pool + "/" + zvolName

	if len(extentDiskName) > 63 {
		return nil, fmt.Errorf("extent zvol name (%s) cannot be longer than 63 chars", extentDiskName)
	}

	if len(zvolName) < 1 {
		return nil, fmt.Errorf("zvol name cannot be empty")
	}

	glog.Infof("Creating target: \"%s\", zvol: \"%s/%s\", extent: \"%s\"", iscsiName, parentDs.Pool, zvolName, iscsiName)

	// Create zvol
	var zvolVolsize int64
	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	zvolVolsize = volSize.Value()
	zvolVolsizeGB := (float64(zvolVolsize) / 1024 / 1024 / 1024)

	zvol := freenas.Zvol{
		Name:        zvolName,
		Comments:    TruncateString(fmt.Sprintf("%s/%s/%s", meta.GetClusterName(), pvcNamespace, pvcName), 1024),
		Compression: config.ZvolCompression, // config - "" (inherit), lz4, gzip-9, etc
		Dedup:       config.ZvolDedup,       // config - nil (inherit), on, off, verify
		Volsize:     strconv.FormatFloat(zvolVolsizeGB, 'f', -1, 64) + " GiB",
		Sparse:      config.ZvolSparse,
		Force:       config.ZvolForce,
		Blocksize:   config.ZvolBlocksize, // config - 512, 1K, 2K, 4K, 8K, 16K, 32K, 64K, 128K
		Dataset:     parentDs,
	}
	resp, err = zvol.Create(freenasServer)
	if err != nil {
		//glog.Infof("zvol error %s", err.Error())
		if resp.StatusCode == 400 && strings.Contains(err.Error(), "dataset already exists") {
			//zvol.Get(freenasServer)
		} else {
			return nil, err
		}
	}

	// Create target
	target := freenas.Target{
		Name:  iscsiName,
		Alias: "",
		Mode:  "iscsi",
	}
	resp, err = target.Create(freenasServer)
	if err != nil {
		// already exists
		if resp.StatusCode != 409 {
			zvol.Delete(freenasServer)
			return nil, err
		}

		target.Get(freenasServer)
	}

	// Create targetgroup(s)
	targetGroup := freenas.TargetGroup{
		Target:         target.ID,
		Authgroup:      config.TargetGroupAuthgroup,
		Authtype:       config.TargetGroupAuthtype,
		Initialdigest:  "Auto",
		Initiatorgroup: config.TargetGroupInitiatorgroup,
		Portalgroup:    config.TargetGroupPortalgroup,
	}
	resp, err = targetGroup.Create(freenasServer)
	if err != nil {
		// cope with craziness
		if resp.StatusCode == 404 {
			loopResp, loopErr = targetGroup.Get(freenasServer)
			if loopErr != nil || loopResp.StatusCode != 200 {
				glog.Infof("failed attempt to create TargetGroup %d", resp.StatusCode)
				target.Delete(freenasServer)
				zvol.Delete(freenasServer)
				return nil, err
			}
		} else if resp.StatusCode != 409 {
			glog.Infof("failed attempt to create TargetGroup %d", resp.StatusCode)
			target.Delete(freenasServer)
			zvol.Delete(freenasServer)
			return nil, err
		}
	}

	// Create extent
	// whole path to zvol Disk including "zvol/" must be <= 63 chars
	extent := freenas.Extent{
		Name:           iscsiName,
		Type:           "Disk",
		Disk:           extentDiskName,
		Blocksize:      config.ExtentBlocksize, //config - 512, 1024, 2048, or 4096
		Pblocksize:     config.ExtentDisablePhysicalBlocksize,
		AvailThreshold: config.ExtentAvailThreshold,
		Comment:        TruncateString(fmt.Sprintf("%s/%s", pvcNamespace, pvcName), 120),
		InsecureTpc:    config.ExtentInsecureTpc,
		Xen:            config.ExtentXen,
		Rpm:            config.ExtentRpm, // config - Unknown, SSD, 5400, 7200, 10000, 15000
		Ro:             config.ExtentReadOnly,
	}

	// run this as a loop as zvol creation returns a 202 and may take a little time to complete
	extentLoopCurrent := 0
	extentMaxLoops := 2
	extentWaitDuration, err := time.ParseDuration("5s")
	for {
		resp, err = extent.Create(freenasServer)
		if err != nil {
			if resp.StatusCode == 409 {
				loopResp, loopErr = extent.Get(freenasServer)
				if loopErr != nil {
					glog.Infof("failed attempt to create Extent %d", resp.StatusCode)
					targetGroup.Delete(freenasServer)
					target.Delete(freenasServer)
					zvol.Delete(freenasServer)
					return nil, err
				}
				break
			}

			if extentMaxLoops == extentLoopCurrent {
				targetGroup.Delete(freenasServer)
				target.Delete(freenasServer)
				zvol.Delete(freenasServer)
				return nil, err
			}
			extentLoopCurrent++
			time.Sleep(extentWaitDuration)
		} else {
			break
		}
	}

	// Create targettoextent
	lunid := 0
	targetToExtent := freenas.TargetToExtent{
		Extent: extent.ID,
		Lunid:  &lunid,
		Target: target.ID,
	}
	resp, err = targetToExtent.Create(freenasServer)
	if err != nil {
		if resp.StatusCode == 409 {
			loopResp, loopErr = targetToExtent.Get(freenasServer)
			if loopErr != nil {
				glog.Infof("failed attempt to create TargetToExtent %d", resp.StatusCode)
				extent.Delete(freenasServer)
				targetGroup.Delete(freenasServer)
				target.Delete(freenasServer)
				zvol.Delete(freenasServer)
				return nil, err
			}
		} else {
			extent.Delete(freenasServer)
			targetGroup.Delete(freenasServer)
			target.Delete(freenasServer)
			zvol.Delete(freenasServer)
			return nil, err
		}
	}

	// use this for testing idempotency
	//return nil, errors.New("fake fail")

	var portals []string
	if len(config.ProvisionerPortals) > 0 {
		portals = strings.Split(config.ProvisionerPortals, ",")
	}

	//mode := v1.PersistentVolumeFilesystem
	//VolumeMode:                    &mode,
	//v1.PersistentVolumeR
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"freenasISCSIProvisionerIdentity": p.Identifier,
				"datasetParent":                   config.DatasetParentName,
				"pool":                            parentDs.Pool,
				"zvol":                            zvolName,
				"iscsiName":                       iscsiName,
				"targetId":                        strconv.Itoa(target.ID),
				"targetGroupId":                   strconv.Itoa(targetGroup.ID),
				"extentId":                        strconv.Itoa(extent.ID),
				"targetToExtentId":                strconv.Itoa(targetToExtent.ID),
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			// set volumeMode from PVC Spec
			VolumeMode: options.PVC.Spec.VolumeMode,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIPersistentVolumeSource{
					TargetPortal:   config.ProvisionerTargetPortal,
					Portals:        portals,
					IQN:            iscsiConfig.Basename + ":" + iscsiName,
					ISCSIInterface: config.ProvisionerISCSIInterface,
					Lun:            int32(*targetToExtent.Lunid),
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
	var targetID, extentID int
	var poolName, zvolName, iscsiName, datasetParentName string

	targetIDAnnotation, ok := volume.Annotations["targetId"]
	if ok {
		targetID, _ = strconv.Atoi(targetIDAnnotation)
	}

	extentIDAnnotation, ok := volume.Annotations["extentId"]
	if ok {
		extentID, _ = strconv.Atoi(extentIDAnnotation)
	}

	poolName = volume.Annotations["pool"]
	zvolName = volume.Annotations["zvol"]
	iscsiName = volume.Annotations["iscsiName"]
	datasetParentName = volume.Annotations["datasetParent"]

	if len(targetIDAnnotation) < 1 {
		return fmt.Errorf("targetID cannot be empty")
	}

	if len(extentIDAnnotation) < 1 {
		return fmt.Errorf("extentID cannot be empty")
	}

	if len(poolName) < 1 {
		return fmt.Errorf("poolName cannot be empty")
	}

	if len(zvolName) < 1 {
		return fmt.Errorf("zvolName cannot be empty")
	}

	if len(datasetParentName) < 1 {
		return fmt.Errorf("datasetParentName cannot be empty")
	}

	var err error
	var resp *http.Response

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
	resp, err = parentDs.Get(freenasServer)
	if err != nil {
		return err
	}

	glog.Infof("Deleting target: %d (\"%s\"), extent: %d (\"%s\"), zvol: \"%s/%s\"", targetID, iscsiName, extentID, iscsiName, poolName, zvolName)

	// Delete target
	// NOTE: deletting a target inherently deletes associated targetgroup(s) and targettoextent(s)
	target := freenas.Target{
		ID: targetID,
	}
	resp, err = target.Delete(freenasServer)
	if err != nil {
		if resp.StatusCode != 404 {
			return err
		}
	}

	// Delete extent
	extent := freenas.Extent{
		ID: extentID,
	}
	resp, err = extent.Delete(freenasServer)
	if err != nil {
		if resp.StatusCode != 404 {
			return err
		}
	}

	// Delete zvol
	zvol := freenas.Zvol{
		Name:    zvolName,
		Dataset: parentDs,
	}
	resp, err = zvol.Delete(freenasServer)
	if err != nil {
		if resp.StatusCode != 404 {
			if resp.StatusCode == 400 && strings.Contains(err.Error(), "dataset does not exist") {
				glog.Infof("Zvol %s/%s already deleted", zvol.Dataset.Name, zvol.Name)
			} else {
				return err
			}
		}
	}

	// use this for testing idempotency
	//return errors.New("fake fail")

	return nil
}

func (p *freenasProvisioner) ShouldProvision(*v1.PersistentVolumeClaim) bool {
	//glog.Infof("ShouldProvision invoked")
	return true
}

func (p *freenasProvisioner) SupportsBlock() bool {
	//glog.Infof("SupportsBlock invoked")
	return true
}

func (p *freenasProvisioner) GetServer(config freenasProvisionerConfig) (*freenas.Server, error) {
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

// BytesToString converts bytes to a string
func BytesToString(data []byte) string {
	return string(data[:])
}

// TruncateString removes all chars from a string after num chars
func TruncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		bnoden = str[:num]
	}
	return bnoden
}

/*
// Prep for resizing
// https://github.com/kubernetes-incubator/external-storage/blob/master/gluster/file/cmd/glusterfile-provisioner/glusterfile-provisioner.go#L433
// https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/d22b74e900af4bf90174d259c3e52c3680b41ab4/controller/volume.go

func (p *freenasProvisioner) RequiresFSResize() bool {
	return true
}

func (p *freenasProvisioner) ExpandVolumeDevice(spec *volume.Spec, newSize resource.Quantity, oldSize resource.Quantity) (resource.Quantity, error) {
	//return newVolumeSize, nil
	glog.Infof("STUFF AND STUFF")
	return oldSize, fmt.Errorf("ExpandVolumeDevice not yet implemented")
}
*/
