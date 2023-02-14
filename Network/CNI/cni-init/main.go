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

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func readFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return content
}

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
				log.Error(err)
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
				panic(err)
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

	//TODO add nat firewall

	log.Info("Init done,bye bye cowboy space")
	time.Sleep(5 * time.Minute)
	os.Exit(0)
}
