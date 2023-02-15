package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-iptables/iptables"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	CNI_CONFIG_PATH  = "/app/config/10-humanz-cni-plugin.conf"
	CNI_BIN_PATH_SRC = "/humanz-cni"
	CNI_BIN_PATH_DST = "/app/bin/humanz-cni"
)

func main() {
	NodeHostName := os.Getenv("HOSTNAME")
	log.WithFields(log.Fields{
		"Hostname": NodeHostName,
	}).Info("Init CNI")

	K8s_SVC := fmt.Sprintf("https://%s/api/v1/nodes/", os.Getenv("KUBERNETES_SERVICE_HOST"))
	token := readFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	ca := readFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	req, err := http.NewRequest("GET", K8s_SVC, nil)
	if err != nil {
		log.Error(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(token)))

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Panic("Invalid status code,check your rbac")
	}

	var NodesInfo k8sNode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)

	}

	err = json.Unmarshal(body, &NodesInfo)
	if err != nil {
		log.Error(err)
	}

	HostPodCIDR := ""
	for _, Node := range NodesInfo.Items {
		if Node.Metadata.Name != NodeHostName {
			//Do ip route
			PodCIDR := Node.Spec.PodCIDR
			NodeIP := func() string {
				for _, v := range Node.Status.Addresses {
					if v.Type == "InternalIP" {
						return v.Address
					}
				}

				return ""
			}()

			_, podNet, err := net.ParseCIDR(PodCIDR)
			if err != nil {
				log.Fatal(err)
			}

			NodeNet := net.ParseIP(NodeIP)
			NewRoute := netlink.Route{
				Gw:  NodeNet,
				Dst: podNet,
			}

			log.WithFields(log.Fields{
				"Hostname":     NodeHostName,
				"Dst Hostname": Node.Metadata.Name,
				"PodCIDR":      PodCIDR,
				"Gateway":      NodeNet,
			}).Info("Setup ip route")

			err = netlink.RouteAdd(&NewRoute)
			if err != nil {
				if err.Error() == "file exists" {
					continue
				}
				log.Fatal(err)
			}
		} else {
			HostPodCIDR = Node.Spec.PodCIDR
		}
	}

	myCni := Humanz_CNI{
		CniVersion: "0.3.1",
		Name:       "humanz-cni",
		Type:       "humanz-cni",
		Bridge:     "humanz-cni0",
		Subnet:     HostPodCIDR,
	}

	log.WithFields(log.Fields{
		"Hostname": NodeHostName,
		"Path":     CNI_CONFIG_PATH,
	}).Info("Dump cni plugin config")

	file, _ := json.MarshalIndent(myCni, "", " ")
	err = ioutil.WriteFile(CNI_CONFIG_PATH, file, 0644)
	if err != nil {
		log.Error(err)
	}

	log.WithFields(log.Fields{
		"src path": CNI_BIN_PATH_SRC,
		"dst path": CNI_BIN_PATH_DST,
	}).Info("Copy cni bin")

	cmd := exec.Command("mv", CNI_BIN_PATH_SRC, CNI_BIN_PATH_DST)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	tab, err := iptables.New()
	if err != nil {
		log.Error(err)
	}

	err = tab.Append("filter", "FORWARD", "-s", HostPodCIDR, "-j", "ACCEPT", "-m", "comment", "--comment", "ACCEPT src pods network")
	if err != nil {
		log.Error(err)
	}

	err = tab.Append("filter", "FORWARD", "-d", HostPodCIDR, "-j", "ACCEPT", "-m", "comment", "--comment", "ACCEPT dst pods network")
	if err != nil {
		log.Error(err)
	}

	NatIface := detectOutsideNat()
	if NatIface == "" {
		log.Warn("Nat to outside network can't be found on all interface,skip the nat")
	} else {
		err = tab.Append("nat", "POSTROUTING", "-s", HostPodCIDR, "-o", NatIface, "-j", "MASQUERADE", "-m", "comment", "--comment", "Nat from pods to outside")
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Init done,bye bye cowboy space")
	time.Sleep(5 * time.Minute)
	os.Exit(0)
}

func readFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return content
}

func detectOutsideNat() string {
	lnkList, err := netlink.LinkList()
	if err != nil {
		log.Error(err)
	}

	InetNatIface := ""
end:
	for _, link := range lnkList {

		linkAddrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			log.Error(err)
		}

		for _, linkAddr := range linkAddrs {
			if err := curlCloudflare(linkAddr.IP); err != nil {
				log.WithFields(log.Fields{
					"Iface": link.Attrs().Name,
					"IP":    linkAddr.IP.String(),
				}).Warn("Nat to outside can't be found : %s", err)
			} else {
				log.WithFields(log.Fields{
					"Iface": link.Attrs().Name,
					"IP":    linkAddr.IP.String(),
				}).Info("Nat to outside found")

				InetNatIface = link.Attrs().Name
				break end
			}
		}
	}

	return InetNatIface
}

func curlCloudflare(ip net.IP) error {
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
