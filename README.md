# Deprecation Notice

Non-csi drivers are no longer receiving updates from the Kubernetes team.
As such this driver has been deprecated and replaced with a newer `csi` based
implementation - https://github.com/democratic-csi/democratic-csi

The new implementation works with all the fancy `csi` features such as snapshots,
resizing, etc. Enjoy!

# What is freenas-provisioner

FreeNAS-iscsi-provisioner is a Kubernetes external provisioner.
When a `PersisitentVolumeClaim` appears on a Kube cluster, the provisioner will
make the corresponding calls to the configured FreeNAS API to create an iscsi
target/lun usable by the claim. When the claim or the persistent volume is
deleted, the provisioner deletes the previously created resources.

See this for more info on external provisioner:
https://github.com/kubernetes-incubator/external-storage

Unless you have a very specific use-case for iscsi/block devices, it is
recommended to use the NFS variant of this project available here:
https://github.com/nmaupu/freenas-provisioner

# Usage

The scope of the provisioner allows for a single instance to service multiple
classes (and/or FreeNAS servers). The provisioner itself can be deployed into
the cluster or ran out of cluster, for example, directly on a FreeNAS server.

Each `StorageClass` should have a corresponding `Secret` created which contains
the credentials and host information used to communicate with with FreeNAS API.
In essence each `Secret` corresponds to a FreeNAS server.

The `Secret` namespace and name may be customized using the appropriate
`StorageClass` `parameters`. By default `kube-system` and `freenas-iscsi` are
used. While multiple `StorageClass` resources may point to the same server
and hence same `Secret`, it is recommended to create a new `Secret` for each
`StorageClass` resource.

It is **highly** recommended to read `deploy/class.yaml` to review available
`parameters` and gain a better understanding of functionality and behavior.

## FreeNAS Setup

You must manually create a dataset. You may simply use a pool as the parent
dataset but it's recommended to create a dedicated dataset.

Additionally, you need to enable the iscsi service with it's corresponding
resources such as portal, initiator, and group.

## Provision the provisioner

Run it on the cluster:

```
kubectl apply -f deploy/rbac.yaml -f deploy/deployment.yaml
```

Alternatively, for advanced use-cases you may run the provisioner out of cluster
including directly on the FreeNAS server if desired. Running out of cluster is
not currently recommended.

```
./bin/freenas-iscsi-provisioner-freebsd --kubeconfig=/path/to/kubeconfig.yaml
```

## Create `StorageClass` and `Secret`

All the necessary resources are available in the `deploy` folder. At a minimum
`secret.yaml` must be modified (remember to `base64` the values) to reflect the
server details. You may also want to read `class.yaml` to review available
`parameters` of the storage class. For instance to set the `datasetParentName`.

```
kubectl apply -f deploy/secret.yaml -f deploy/class.yaml
```

## Example usage

Next, create a `PersistentVolumeClaim` using the storage class
(`deploy/test-claim.yaml`):

```
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: freenas-test-iscsi-pvc
spec:
  storageClassName: freenas-iscsi
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Mi
```

Use that claim on a testing pod (`deploy/test-pod.yaml`):

```
---
kind: Pod
apiVersion: v1
metadata:
  name: freenas-test-iscsi-pod
spec:
  containers:
  - name: freenas-test-isci-pod
    image: gcr.io/google_containers/busybox:1.24
    command:
      - "/bin/sh"
      - "-c"
      - "--"
    args: [ "date >> /mnt/file.log && while true; do sleep 30; done;" ]
    volumeMounts:
      - name: freenas-test-volume
        mountPath: "/mnt"
  restartPolicy: "Never"
  volumes:
    - name: freenas-test-volume
      persistentVolumeClaim:
        claimName: freenas-test-iscsi-pvc
```

The underlying zvol, target, extent, etc should be quickly appearing on the
FreeNAS side. In case of issue, follow the provisioner's logs using:

```
kubectl -n kube-system logs -f freenas-iscsi-provisioner-<id>
```

## CHAP settings

You should create a secret which holds CHAP authentication credentials based on `deploy/freenas-iscsi-chap.yaml`.
- If you have authentication enabled for the portal (discovery) then set `discovery*` parameters in the secret, and in StorageClass you should set `targetDiscoveryCHAPAuth` to `true`.
- If you want authentication for the targets, then set `node*` parameters in the secret, and in StorageClass you should set `targetGroupAuthtype` and `targetGroupAuthgroup` accordingly, and also set `targetSessionCHAPAuth` to `true`.

# Performance

100 10MiB PVCs
Creating took ~10 minutes

Deleting took ~6 minutes

# Testing

Choas testing has been performed to ensure the various actions are idempotent.

# Development

```
make vendor && make
```

Binary is located into `bin/freenas-iscis-provisioner`. It is compiled to be
run on `linux-amd64` by default, but you may run the following for different
builds:

```
make vendor && make darwin
# OR
make vendor && make freebsd
```

To run locally with an appropriate `$KUBECONFIG` you may run:

```
./local-start.sh
```

To format code before committing:

```
make fmt
```

## Docs

- https://github.com/kubernetes/community/tree/master/contributors/design-proposals/storage
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-provisioning.md
- https://kubernetes.io/docs/concepts/storage/storage-classes/
- https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#-strong-api-overview-strong-
- https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_iscsi.html
- http://api.freenas.org
- https://doc.freenas.org/11/sharing.html#block-iscsi
- https://github.com/kubernetes-incubator/external-storage/blob/master/iscsi/targetd/provisioner/iscsi-provisioner.go
- https://github.com/dghubble/sling

## TODO

- volume resizing - https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/grow-volume-size.md
- volume snapshots - https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-snapshotting.md
- mount options - https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/mount-options.md
- ~~CHAP~~
- fsType
- properly handle `zvol` API differences with `volsize` getting sent as string and returned as int
- loop GetBy<foo> requests that require `limit` param
- ~~recursive zvol delete in v1 api~~

## Notes

To sniff API traffic between host and server:

```
sudo tcpdump -i any -A -s 0 'host <server ip> and tcp port 80 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'
```
