package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	pluginName = "humanz_cni"
	LOG_FILE   = "/var/log/humanz_cni.log"
	IP_STORAGE = "/run/humanz-cni.json"
	MTU        = 1500
)

type Humanz_CNI struct {
	CniVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Bridge     string `json:"bridge"`
	Network    string `json:"network"`
	Subnet     string `json:"subnet"`
}

type Humanz_CNI_Storage struct {
	IP   string `json:"IP"`
	Used bool   `json:"Used"`
}

func CreateBridge(bridge string, mtu int, gateway *net.IPNet) (netlink.Link, error) {
	if l, _ := netlink.LinkByName(bridge); l != nil {
		return l, nil
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   bridge,
			MTU:    mtu,
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

var (
	logger = log.New()
)

func init() {
	runtime.LockOSThread()
}

func main() {
	logger.SetOutput(os.Stdout)

	file, err := os.OpenFile(LOG_FILE, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()
	logger.SetOutput(file)

	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString(pluginName))
}

func cmdAdd(args *skel.CmdArgs) error {
	sb := Humanz_CNI{}
	err := json.Unmarshal(args.StdinData, &sb)
	if err != nil {
		return err
	}

	IPS := InitDB(sb.Subnet)
	BridgeName := sb.Bridge
	ip, GatewayIP, err := net.ParseCIDR(IPS[0].IP)
	if err != nil {
		return err
	}

	Gate := net.IPNet{
		IP:   ip,
		Mask: GatewayIP.Mask,
	}

	br, err := CreateBridge(BridgeName, MTU, &Gate)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return err
	}
	defer netns.Close()

	PodIP, err := GeneratePodsIP()
	if err != nil {
		return err
	}

	err = SetupVeth(netns, br, args.IfName, PodIP, ip)
	if err != nil {
		return err
	}

	logger.WithFields(log.Fields{
		"NS":           netns.Path(),
		"PodIP":        PodIP.String(),
		"Container ID": args.ContainerID,
	}).Info("New pod created")

	result := &current.Result{
		CNIVersion: sb.CniVersion,
		IPs: []*current.IPConfig{
			{
				Address: *PodIP,
				Gateway: ip,
			},
		},
	}

	return types.PrintResult(result, sb.CniVersion)

}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return err
	}
	defer netns.Close()

	err = netns.Do(func(nn ns.NetNS) error {
		NsNet, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return err
		}

		NsNetAddr, err := netlink.AddrList(NsNet, netlink.FAMILY_V4)
		if err != nil {
			return err
		}

		for _, v := range NsNetAddr {
			if v.IPNet != nil {
				logger.WithFields(log.Fields{
					"IP":           v.IPNet.String(),
					"NS":           args.Netns,
					"Container ID": args.ContainerID,
				}).Info("Deleting ip from db")
				err := RemoveIP(v.IPNet)
				if err != nil {
					return err
				}
			}
		}

		err = netns.Close()
		return err
	})

	return err
}
