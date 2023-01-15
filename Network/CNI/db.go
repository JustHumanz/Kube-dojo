package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func InitDB(cidr string) []Humanz_CNI_Storage {
	var IPS []Humanz_CNI_Storage
	if _, err := os.Stat(IP_STORAGE); err == nil {
		IPS = ReadDB()
	} else {
		IPS = CountIP(cidr)
		err := UpdateStorage(IPS)
		if err != nil {
			panic(err)
		}
	}

	return IPS
}

func ReadDB() []Humanz_CNI_Storage {
	jsonFile, err := os.Open(IP_STORAGE)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	var (
		IPS []Humanz_CNI_Storage
	)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &IPS)
	if err != nil {
		panic(err)
	}

	return IPS
}

func CountIP(cidr string) []Humanz_CNI_Storage {
	var IPS []Humanz_CNI_Storage

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	ipmasklen, _ := ipnet.Mask.Size()
	total := 1<<uint(net.IPv4len*8-ipmasklen) - 2

	for z := 1; z <= total; z++ {
		counter := z
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		for i := len(newIP) - 1; i >= 0; i-- {
			newIP[i] += byte(counter)
			counter >>= 8
			if counter == 0 {
				break
			}
		}

		tmp := Humanz_CNI_Storage{
			IP:   fmt.Sprintf("%s/%d", newIP.String(), ipmasklen),
			Used: false,
		}

		if z == 1 {
			tmp.Used = true
		}

		IPS = append(IPS, tmp)
	}

	return IPS
}

func UpdateStorage(DataIP []Humanz_CNI_Storage) error {
	file, err := json.MarshalIndent(DataIP, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(IP_STORAGE, file, 0644)
	return err
}

func GeneratePodsIP() (*net.IPNet, error) {
	var (
		IPS   = ReadDB()
		PodIP net.IPNet
	)

	for i, v := range IPS {
		if !v.Used {
			ip, ipnet, err := net.ParseCIDR(v.IP)
			if err != nil {
				return nil, err
			}

			PodIP = net.IPNet{
				IP:   ip,
				Mask: ipnet.Mask,
			}

			IPS[i].Used = true
			break
		}
	}

	UpdateStorage(IPS)

	return &PodIP, nil
}

func RemoveIP(DelIP *net.IPNet) error {
	var (
		IPS = ReadDB()
	)

	for i, v := range IPS {
		if v.Used {
			ip, _, err := net.ParseCIDR(v.IP)
			if err != nil {
				return nil
			}

			if ip.Equal(DelIP.IP) {
				IPS[i].Used = false
				break
			}
		}
	}

	UpdateStorage(IPS)

	return nil
}
