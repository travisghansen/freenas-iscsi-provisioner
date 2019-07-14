# to watch after creation
watch kubectl -n freenas-iscsi-test get pvc

# to remove them all
kubectl -n freenas-iscsi-test delete pvc --all
