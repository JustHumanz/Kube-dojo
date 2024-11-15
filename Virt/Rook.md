## Create volume
- openstack volume create rook-1
- openstack volume create rook-2
- openstack volume create rook-3
- openstack server add volume worker-1 rook-1
- openstack server add volume worker-2 rook-2
- openstack server add volume worker-3 rook-3

## Install rook
- git clone --single-branch --branch master https://github.com/rook/rook.git
- cd rook/deploy/examples
- kubectl create -f crds.yaml -f common.yaml -f operator.yaml
- nano cluster.yaml
- kubectl create -f cluster.yaml toolbox.yaml
- nano filesystem.yaml
- kubectl create -f filesystem.yaml
- nano csi/cephfs/storageclass.yaml
- kubectl create -f csi/cephfs/storageclass.yaml

## Tshoot
- kubectl -n rook-ceph delete pod -l app=rook-ceph-operator
