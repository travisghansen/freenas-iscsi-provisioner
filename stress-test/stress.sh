#!/bin/bash


for i in {1..50}
do

cat <<EOF | kubectl create -f -
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: freenas-test-iscsi-pvc-${i}
  namespace: freenas-iscsi-test
spec:
  storageClassName: freenas-iscsi
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Mi
EOF

done



