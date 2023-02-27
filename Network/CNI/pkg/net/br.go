package iface

import (
	"fmt"
	"net"
	"syscall"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

const MTU = 1500

func CreateBridge(bridge string, gateway *net.IPNet) (netlink.Link, error) {
	if l, _ := netlink.LinkByName(bridge); l != nil {
		return l, nil
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   bridge,
			MTU:    MTU,
			TxQLen: -1,
		},
	}

	if err := netlink.LinkAdd(br); err != nil && err != syscall.EEXIST {
		return nil, err
	}

	dev, err := netlink.LinkByName(bridge)
	if err != nil {
		return nil, err
	}

	if err := netlink.AddrAdd(dev, &netlink.Addr{
		IPNet: gateway,
	}); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(dev); err != nil {
		return nil, err
	}

	return dev, nil
}

func SetupVeth(netns ns.NetNS, br netlink.Link, ifName string, podIP *net.IPNet, gateway net.IP) error {
	hostIface := &current.Interface{}
	err := netns.Do(func(hostNS ns.NetNS) error {
		// setup lo, kubernetes will call loopback internal
		// loLink, err := netlink.LinkByName("lo")
		// if err != nil {
		// 	return err
		// }

		// if err := netlink.LinkSetUp(loLink); err != nil {
		// 	return err
		// }

		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, MTU, "", hostNS)
		if err != nil {
			return err
		}

		hostIface.Name = hostVeth.Name

		// set ip for container veth
		conLink, err := netlink.LinkByName(containerVeth.Name)
		if err != nil {
			return err
		}
		if err := netlink.AddrAdd(conLink, &netlink.Addr{IPNet: podIP}); err != nil {
			return err
		}

		// setup container veth
		if err := netlink.LinkSetUp(conLink); err != nil {
			return err
		}

		// add default route
		if err := ip.AddDefaultRoute(gateway, conLink); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// need to lookup hostVeth again as its index has changed during ns move
	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return fmt.Errorf("failed to lookup %q: %v", hostIface.Name, err)
	}

	if hostVeth == nil {
		return fmt.Errorf("nil hostveth")
	}

	// connect host veth end to the bridge
	if err := netlink.LinkSetMaster(hostVeth, br); err != nil {
		return fmt.Errorf("failed to connect %q to bridge %v: %v", hostVeth.Attrs().Name, br.Attrs().Name, err)
	}

	return nil
}
