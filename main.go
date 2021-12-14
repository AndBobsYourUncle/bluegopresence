package main

import (
	"fmt"
	"io/ioutil"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"gopkg.in/yaml.v2"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

type KnownDevice struct {
	MacAddress string `yaml:"mac_address"`
	Name       string `yaml:"name"`
}

type KnownStaticDevices struct {
	Devices []KnownDevice `yaml:"devices"`
}

type MqttConfig struct {
	Host                    string `yaml:"host"`
	Port                    int    `yaml:"port"`
	Username                string `yaml:"username"`
	Password                string `yaml:"password"`
	Topic                   string `json:"topic"`
	PublishFrequencySeconds int    `json:"publish_frequency_seconds"`
}

type KnownDeviceTracking struct {
	Name        string
	LastPublish time.Time
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

	deviceMap := make(map[string]*KnownDeviceTracking)

	for _, knownStaticDevice := range knownStaticDevices.Devices {
		deviceMap[knownStaticDevice.MacAddress] = &KnownDeviceTracking{Name: knownStaticDevice.Name}
	}

	body, err = ioutil.ReadFile("./mqtt_config.yaml")
	if err != nil {
		panic("failed to open mqtt_config.yaml")
	}

	mqttConfig := MqttConfig{}

	err = yaml.Unmarshal(body, &mqttConfig)
	if err != nil {
		panic("error parsing mqtt_config.yaml")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttConfig.Host, mqttConfig.Port))
	opts.SetClientID("bluegopresence")
	opts.SetUsername(mqttConfig.Username)
	opts.SetPassword(mqttConfig.Password)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	defer adapter.StopScan()

	defer client.Disconnect(0)

	// Start scanning.
	println("scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if knownDevice, ok := deviceMap[device.Address.String()]; ok {
			if time.Now().After(knownDevice.LastPublish.Add(time.Second * time.Duration(mqttConfig.PublishFrequencySeconds))) {
				println(time.Now().String())
				println(knownDevice.LastPublish.String())

				knownDevice.LastPublish = time.Now()

				text := fmt.Sprintf("%d", device.RSSI)
				token := client.Publish("bluegopresence/"+mqttConfig.Topic+"/"+knownDevice.Name+"/rssi", 0, false, text)
				token.Wait()
			}
		}
	})
	must("start scan", err)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
