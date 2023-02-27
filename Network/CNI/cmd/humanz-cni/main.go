package main

import (
	"encoding/json"
	"net"
	"os"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/justhumanz/humanz-cni/pkg/db"
	iface "github.com/justhumanz/humanz-cni/pkg/net"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	pluginName = "humanz_cni"
	LOG_FILE   = "/var/log/humanz_cni.log"
)

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
	sb := db.Humanz_CNI{}
	err := json.Unmarshal(args.StdinData, &sb)
	if err != nil {
		return err
	}

	IPS := db.InitDB(sb.Subnet)
	BridgeName := sb.Bridge
	gatewayIP, gatewayNet, err := net.ParseCIDR(IPS[0].IP)
	if err != nil {
		return err
	}

	Gate := net.IPNet{
		IP:   gatewayIP,
		Mask: gatewayNet.Mask,
	}

	br, err := iface.CreateBridge(BridgeName, &Gate)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return err
	}
	defer netns.Close()

	PodIP, err := db.GeneratePodsIP()
	if err != nil {
		return err
	}

	err = iface.SetupVeth(netns, br, args.IfName, PodIP, gatewayIP)
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
				Gateway: gatewayIP,
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
				err := db.RemoveIP(v.IPNet)
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
