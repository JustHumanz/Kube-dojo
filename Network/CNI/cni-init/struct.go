package main

import "time"

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

type Humanz_CNI struct {
	CniVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Bridge     string `json:"bridge"`
	Subnet     string `json:"subnet"`
}
