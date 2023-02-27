package iface

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/vishvananda/netlink"
)

func DetectOutsideNat() string {
	lnkList, err := netlink.LinkList()
	if err != nil {
		log.Panic(err)
	}

	InetNatIface := ""
end:
	for _, link := range lnkList {

		linkAddrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			log.Panic(err)
		}

		for _, linkAddr := range linkAddrs {
			if err := CurlCloudflare(linkAddr.IP); err == nil {
				InetNatIface = link.Attrs().Name
				break end
			}
		}
	}

	return InetNatIface
}

func CurlCloudflare(ip net.IP) error {
	localTCPAddr := net.TCPAddr{
		IP: ip,
	}

	webclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: &localTCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	req, err := http.NewRequest("GET", "http://1.1.1.1", nil)
	if err != nil {
		return err
	}

	resp, err := webclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil

}
