# Define required providers
terraform {
required_version = ">= 0.14.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.53.0"
    }
  }
}

# Configure the OpenStack Provider
provider "openstack" {  
  user_name   = "${var.username}"
  tenant_name = "${var.project}"
  password    = "${var.password}"
  auth_url    = "https://auth.vexxhost.net/"
  region      = "${var.region}"
}


variable "vms" {
  type = map(any)
  default = {
    ctl1 = {
      ip_address = "10.10.10.11"
    }
    ctl2 = {
      ip_address = "10.10.10.12"
    }
    ctl3 = {
      ip_address = "10.10.10.13"
    }
  }
}

resource "openstack_networking_router_v2" "router1" {
  name           = "router1"
  admin_state_up = "true"
  external_network_id = var.extNet_ID
}

resource "openstack_networking_network_v2" "network_internal" {
  name           = "internal"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "internal-subnet-1" {
  name       = "internal-subnet-1"
  network_id = openstack_networking_network_v2.network_internal.id
  cidr       = "10.10.10.0/24"
  ip_version = 4
}

resource "openstack_networking_router_interface_v2" "int_1" {
  router_id = openstack_networking_router_v2.router1.id
  subnet_id = openstack_networking_subnet_v2.internal-subnet-1.id
}

resource "openstack_networking_port_v2" "ports" {
  for_each       = var.vms
  name           = "${each.key}-port"
  network_id     = openstack_networking_network_v2.network_internal.id
  admin_state_up = "true"
  
  allowed_address_pairs {
    ip_address = "10.10.10.100"
  }

  allowed_address_pairs {
    ip_address = "10.10.10.101"
  }

  fixed_ip {
    subnet_id  = openstack_networking_subnet_v2.internal-subnet-1.id
    ip_address = each.value.ip_address
  }
}


resource "openstack_networking_floatingip_v2" "floatip_1" {
  pool = "${var.extNet_Name}"
}

resource "openstack_networking_floatingip_associate_v2" "ctl1_fip" {
  floating_ip = openstack_networking_floatingip_v2.floatip_1.address
  port_id     = openstack_networking_port_v2.ports["ctl1"].id
}

resource "openstack_compute_instance_v2" "ctl" {

  for_each        = var.vms
  name            = each.key
  flavor_id       = "${var.flavor}"
  key_pair        = "${var.keypair}"
  security_groups = ["default"]
  user_data       = templatefile("scripts/patch_hosts.sh",{domain = var.domain})

  block_device {
    uuid                  = "fe32ec51-1e75-4500-8b81-ba385889b9ec"
    source_type           = "image"
    volume_size           = 50
    boot_index            = 0
    destination_type      = "volume"
    delete_on_termination = true
  }

  block_device {
    source_type           = "blank"
    destination_type      = "volume"
    volume_size           = 80
    boot_index            = 1
    delete_on_termination = true
  }

  network {
    port = openstack_networking_port_v2.ports[each.key].id
  }
}

