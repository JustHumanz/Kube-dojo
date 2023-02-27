package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/justhumanz/humanz-cni/pkg/db"
	iface "github.com/justhumanz/humanz-cni/pkg/net"
	log "github.com/sirupsen/logrus"
	k8sv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	HostPodCIDR := ""
	for _, Node := range nodeList.Items {
		if Node.Name != NodeHostName {
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

			err := iface.AddNewRoute(PodCIDR, NodeIP)
			if err != nil {
				log.Panic(err)
			}

		} else {
			HostPodCIDR = Node.Spec.PodCIDR
		}
	}

	myCni := db.Humanz_CNI{
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
	err = ioutil.WriteFile(CNI_CONFIG_PATH, file, 0755)
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

	err = tab.AppendUnique("filter", "FORWARD", "-s", HostPodCIDR, "-j", "ACCEPT", "-m", "comment", "--comment", "ACCEPT src pods network")
	if err != nil {
		log.Error(err)
	}

	err = tab.AppendUnique("filter", "FORWARD", "-d", HostPodCIDR, "-j", "ACCEPT", "-m", "comment", "--comment", "ACCEPT dst pods network")
	if err != nil {
		log.Error(err)
	}

	NatIface := iface.DetectOutsideNat()
	if NatIface == "" {
		log.Warn("Nat to outside network can't be found on all interface,skip the nat")
	} else {
		err = tab.AppendUnique("nat", "POSTROUTING", "-s", HostPodCIDR, "-o", NatIface, "-j", "MASQUERADE", "-m", "comment", "--comment", "Nat from pods to outside")
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Init done,bye bye cowboy space")

	CniNodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	knodeList := make(map[string]bool)

	for _, v := range CniNodeList.Items {
		knodeList[v.Name] = true
	}

	NodesWatch, err := clientset.CoreV1().Nodes().Watch(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for NodesEvent := range NodesWatch.ResultChan() {
		Node := NodesEvent.Object.(*k8sv1.Node)
		if !knodeList[Node.Name] {

			newNode, err := clientset.CoreV1().Nodes().Get(context.TODO(), Node.Name, v1.GetOptions{})
			if err != nil {
				log.Fatal(err)
			}

			PodCIDR := newNode.Spec.PodCIDR
			NodeIP := func() string {
				for _, v := range newNode.Status.Addresses {
					if v.Type == "InternalIP" {
						return v.Address
					}
				}

				return ""
			}()

			log.WithFields(log.Fields{
				"NodeName": Node.Name,
				"PodsCIDR": PodCIDR,
				"NodeIP":   NodeIP,
			}).Info("New node join")

			//Add ip route to new node
			err = iface.AddNewRoute(PodCIDR, NodeIP)
			if err != nil {
				log.Fatal(err)
			}

			knodeList[Node.Name] = true
		}
	}

	os.Exit(0)
}

func readFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return content
}

type k8sNode struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				BetaKubernetesIoArch             string `json:"beta.kubernetes.io/arch"`
				BetaKubernetesIoOs               string `json:"beta.kubernetes.io/os"`
				KubernetesIoArch                 string `json:"kubernetes.io/arch"`
				KubernetesIoHostname             string `json:"kubernetes.io/hostname"`
				KubernetesIoOs                   string `json:"kubernetes.io/os"`
				NodeRoleKubernetesIoControlPlane string `json:"node-role.kubernetes.io/control-plane"`
				NodeRoleKubernetesIoMaster       string `json:"node-role.kubernetes.io/master"`
			} `json:"labels"`
			Annotations struct {
				KubeadmAlphaKubernetesIoCriSocket                string `json:"kubeadm.alpha.kubernetes.io/cri-socket"`
				NodeAlphaKubernetesIoTTL                         string `json:"node.alpha.kubernetes.io/ttl"`
				ProjectcalicoOrgIPv4Address                      string `json:"projectcalico.org/IPv4Address"`
				ProjectcalicoOrgIPv4IPIPTunnelAddr               string `json:"projectcalico.org/IPv4IPIPTunnelAddr"`
				VolumesKubernetesIoControllerManagedAttachDetach string `json:"volumes.kubernetes.io/controller-managed-attach-detach"`
			} `json:"annotations"`
		} `json:"metadata"`
		Spec struct {
			PodCIDR  string   `json:"podCIDR"`
			PodCIDRs []string `json:"podCIDRs"`
		} `json:"spec"`
		Status struct {
			Capacity struct {
				CPU              string `json:"cpu"`
				EphemeralStorage string `json:"ephemeral-storage"`
				Hugepages1Gi     string `json:"hugepages-1Gi"`
				Hugepages2Mi     string `json:"hugepages-2Mi"`
				Memory           string `json:"memory"`
				Pods             string `json:"pods"`
			} `json:"capacity"`
			Allocatable struct {
				CPU              string `json:"cpu"`
				EphemeralStorage string `json:"ephemeral-storage"`
				Hugepages1Gi     string `json:"hugepages-1Gi"`
				Hugepages2Mi     string `json:"hugepages-2Mi"`
				Memory           string `json:"memory"`
				Pods             string `json:"pods"`
			} `json:"allocatable"`
			Conditions []struct {
				Type               string    `json:"type"`
				Status             string    `json:"status"`
				LastHeartbeatTime  time.Time `json:"lastHeartbeatTime"`
				LastTransitionTime time.Time `json:"lastTransitionTime"`
				Reason             string    `json:"reason"`
				Message            string    `json:"message"`
			} `json:"conditions"`
			Addresses []struct {
				Type    string `json:"type"`
				Address string `json:"address"`
			} `json:"addresses"`
			DaemonEndpoints struct {
				KubeletEndpoint struct {
					Port int `json:"Port"`
				} `json:"kubeletEndpoint"`
			} `json:"daemonEndpoints"`
			NodeInfo struct {
				MachineID               string `json:"machineID"`
				SystemUUID              string `json:"systemUUID"`
				BootID                  string `json:"bootID"`
				KernelVersion           string `json:"kernelVersion"`
				OsImage                 string `json:"osImage"`
				ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
				KubeletVersion          string `json:"kubeletVersion"`
				KubeProxyVersion        string `json:"kubeProxyVersion"`
				OperatingSystem         string `json:"operatingSystem"`
				Architecture            string `json:"architecture"`
			} `json:"nodeInfo"`
			Images []struct {
				Names     []string `json:"names"`
				SizeBytes int      `json:"sizeBytes"`
			} `json:"images"`
		} `json:"status"`
	} `json:"items"`
}
