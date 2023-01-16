## Bird protocol

### Ubuntu-nested-2
- `ip netns add humanz_1`
- `ip link add dev veth_humanz_1 type veth peer name eth0 netns humanz_1`
- `ip link set dev veth_humanz_1 up`
- `ip netns exec humanz_1 ip link set dev lo up`
- `ip netns exec humanz_1 ip link set dev eth0 up`
- `ip netns exec humanz_1 ip address add 20.0.1.10 dev eth0`
- `ip netns exec humanz_1 ip route add default via 20.0.1.10`
- `sysctl --write net.ipv4.conf.veth_humanz_1.proxy_arp=1`
- `ip route add 20.0.1.10/32 dev veth_humanz_1`

### Ubuntu-nested-3
- `ip netns add humanz_2`
- `ip link add dev veth_humanz_2 type veth peer name eth0 netns humanz_2`
- `ip link set dev veth_humanz_2 up`
- `ip netns exec humanz_2 ip link set dev lo up`
- `ip netns exec humanz_2 ip link set dev eth0 up`
- `ip netns exec humanz_2 ip address add 20.0.2.10 dev eth0`
- `ip netns exec humanz_2 ip route add default via 20.0.2.10`
- `sysctl --write net.ipv4.conf.veth_humanz_2.proxy_arp=1`
- `ip route add 20.0.2.10/32 dev veth_humanz_2`

### Both VM
- `sudo sysctl --write net.ipv4.ip_forward=1`


### Ubuntu-nested-2
- `ip netns exec humanz_1 ping 20.0.2.10`

### Ubuntu-nested-3
- `ip netns exec humanz_2 ping 20.0.1.10`


## Setup Bird
- `sudo apt update && sudo apt install bird2 --yes` # Both vm
- `nano /etc/bird/bird.conf`
```
log syslog all;

router id 200.0.0.XX;

protocol device {
}

protocol direct {
  ipv4;
}

protocol kernel {
  ipv4 {
    export all;
  };
}

protocol static {
  ipv4;
  route 20.0.X.0/24 blackhole; 
}

protocol bgp ubuntu-nested-Y {
  local 200.0.0.XX as 65000;
  neighbor 200.0.0.YY as 65000;

  ipv4 {
    import all;
    export all;
  };
}

```

X is represent the vm,if that config in ubuntu-nested-2 then use ubuntu-nested-2 vm 
Y is reversal from X

- `systemctl restart bird`
- `birdc show protocols all`
- `ip route`

