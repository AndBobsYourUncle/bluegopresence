package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

type KnownDevice struct {
	MacAddress string `yaml:"mac_address"`
}

type KnownStaticDevices struct {
	Devices []KnownDevice `yaml:"devices"`
}

func main() {
	body, err := ioutil.ReadFile("./known_static_devices.yaml")
	if err != nil {
		panic("failed to open known_static_devices.yaml")
	}

	knownStaticDevices := KnownStaticDevices{}

	err = yaml.Unmarshal(body, &knownStaticDevices)
	if err != nil {
		panic("error parsing known_static_devices.yaml")
	}

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	println("scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	})
	must("start scan", err)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
