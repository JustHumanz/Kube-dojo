package iface

import (
	"net"

	"github.com/vishvananda/netlink"
)

func AddNewRoute(PodCIDR, NodeIP string) error {
	_, podNet, err := net.ParseCIDR(PodCIDR)
	if err != nil {
		return err
	}

	NodeNet := net.ParseIP(NodeIP)
	NewRoute := netlink.Route{
		Gw:  NodeNet,
		Dst: podNet,
	}

	//log.WithFields(log.Fields{
	//	"PodCIDR": PodCIDR,
	//	"Gateway": NodeNet,
	//}).Info("Setup ip route")

	err = netlink.RouteAdd(&NewRoute)
	if err != nil {
		if err.Error() == "file exists" {
			return nil
		}

		return err
	}

	return nil
}
