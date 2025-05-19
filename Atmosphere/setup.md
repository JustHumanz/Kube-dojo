- `apt update;apt install python3-venv python3-dev gcc -y`
- `python3 -m venv atmosphere-env`
- `source ~/atmosphere-env/bin/activate`
- `pip install ansible`
- `mkdir -p atmosphere-cloud-config`
- `nano atmosphere-cloud-config/requirements.yaml`
```yml
---
collections:
  - name: vexxhost.atmosphere
    version: 4.5.0
```
- `ansible-galaxy collection install -r atmosphere-cloud-config/requirements.yaml`
- `mkdir -p atmosphere-cloud-config/playbooks`
- `ansible-playbook -e "workspace_path=~/atmosphere-cloud-config/inventory" -e "ceph_public_network=10.10.10.0/24" -e "domain_name=humanz.cloud" vexxhost.atmosphere.generate_workspace`
- `cd ~/atmosphere-cloud-config/`
- `nano inventory/hosts.ini`
```yaml
[controllers]
ctl1.humanz.cloud
ctl2.humanz.cloud
ctl3.humanz.cloud

[cephs]
ceph1.humanz.cloud
ceph2.humanz.cloud
ceph3.humanz.cloud

[computes]
kvm1.humanz.cloud
kvm2.humanz.cloud
kvm3.humanz.cloud
```
- `nano inventory/group_vars/cephs/osds.yml`
```yaml
ceph_osd_devices:
- /dev/vdb
```
- `nano inventory/group_vars/all/ceph.yml`
```yaml
ceph_mon_fsid: d6c9dd7d-9fa6-5eae-997e-a7d8350c8449
ceph_mon_public_network: 10.10.10.0/24 #ip a 
```
- `nano inventory/group_vars/all/endpoints.yml`
```yaml
keycloak_host: keycloak.humanz.cloud
kube_prometheus_stack_alertmanager_host: alertmanager.humanz.cloud
kube_prometheus_stack_grafana_host: grafana.humanz.cloud
kube_prometheus_stack_prometheus_host: prometheus.humanz.cloud
....
......
.......
........
.........
```
- `nano inventory/group_vars/all/neutron.yml`
```yaml
neutron_networks:
- external: true
  mtu_size: 1500
  name: public
  port_security_enabled: true
  provider_network_type: flat
  provider_physical_network: external
  shared: true
  subnets:
  - allocation_pool_end: 192.168.100.250
    allocation_pool_start: 192.168.100.100
    cidr: 192.168.100.0/24
    enable_dhcp: true
    gateway_ip: 192.168.100.254
    name: public-subnet
```
- `nano inventory/group_vars/all/keepalived.yml`
```yaml
keepalived_interface: ens3
keepalived_vip: 10.10.10.100
```
- `nano inventory/group_vars/all/kubernetes.yml`
```yaml
kubernetes_hostname: k8s.humanz.cloud
kubernetes_keepalived_vip: 10.10.10.101
kubernetes_keepalived_vrid: 42
kubernetes_keepalived_interface: ens3
```
- `nano inventory/group_vars/all/all.yml`
```yaml
---
ovn_helm_values:
  conf:
    auto_bridge_add:
      br-ex: bond0
    ovn_bridge_mappings: external:br-ex
    ovn_bridge_datapath_type: netdev
    
cluster_issuer_type: self-signed
csi_driver: rbd
atmosphere_network_backend: ovn

barbican_helm_values:
  pod:
    replicas:
      api: 1

glance_helm_values:
  pod:
    replicas:
      api: 1
glance_images:
  - name: cirros
    url: http://download.cirros-cloud.net/0.6.2/cirros-0.6.2-x86_64-disk.img
    min_disk: 1
    disk_format: raw
    container_format: bare
    is_public: true

cinder_helm_values:
  pod:
    replicas:
      api: 1
      scheduler: 1

placement_helm_values:
  pod:
    replicas:
      api: 1

nova_helm_values:
  pod:
    replicas:
      api_metadata: 1
      osapi: 1
      conductor: 1
      novncproxy: 1
      spiceproxy: 1
  conf:
    nova:
      DEFAULT:
        osapi_compute_workers: 2
        metadata_workers: 2
      conductor:
        workers: 2
      scheduler:
        workers: 2
        
neutron_helm_values:
  conf:
    neutron:
      DEFAULT:
        api_workers: 2
        rpc_workers: 2
        metadata_workers: 2

heat_helm_values:
  conf:
    heat:
      DEFAULT:
        num_engine_workers: 2
      heat_api:
        workers: 2
      heat_api_cfn:
        workers: 2
      heat_api_cloudwatch:
        workers: 2
  pod:
    replicas:
      api: 1
      cfn: 1
      cloudwatch: 1
      engine: 1

octavia_helm_values:
  conf:
    octavia:
      controller_worker:
        workers: 2
    octavia_api_uwsgi:
      uwsgi:
        processes: 2
  pod:
    replicas:
      api: 1
      worker: 1
      housekeeping: 1              
```
- `nano ../playbooks/site.yml`
```yaml
---
- import_playbook: vexxhost.atmosphere.site
```

- `ansible-playbook -i inventory/hosts.ini playbooks/site.yml` # or `ansible-playbook -i inventory/hosts.ini vexxhost.atmosphere.ceph X X X`



### Misc
`ansible -i inventory/hosts.ini all -m raw -a "resolvectl dns ens3 1.1.1.1"`
`kubectl get ingress -A`
`XXXXXX keycloak.humanz.cloud   alertmanager.humanz.cloud       grafana.humanz.cloud    prometheus.humanz.cloud key-manager.humanz.cloud        volume.humanz.cloud     dns.humanz.cloud        image.humanz.cloud      orchestration.humanz.cloud      cloudformation.humanz.cloud     dashboard.humanz.cloud  baremetal.humanz.cloud identity.humanz.cloud    container-infra.humanz.cloud    container-infra-registry.humanz.cloud   share.humanz.cloud      network.humanz.cloud    compute.humanz.cloud    vnc.humanz.cloud        load-balancer.humanz.cloud      placement.humanz.cloud  object-store.humanz.cloud`
