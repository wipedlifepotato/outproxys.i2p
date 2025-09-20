package models

import "os"
import "encoding/json"
import "outproxys/proxy"
import "log"
import "github.com/spf13/viper"

var dataFile = ""
const defDataFile = "proxies.json"

type Proxy struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Uptime  int    `json:"uptime"`
	Type    string `json:"type"` // "http", "https", "socks"
}

func LoadProxies() ([]Proxy, error) {
	dataFile = viper.GetString("dataFile")
	if dataFile == "" {
		log.Println("Data file is nil")
		dataFile=defDataFile
	}
	log.Println("load proxies from " + dataFile)
	file, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Proxy{}, nil
		}
		return nil, err
	}
	var proxies []Proxy
	err = json.Unmarshal(file, &proxies)
	return proxies, err
}

func SaveProxies(proxies []Proxy) error {
	bytes, err := json.MarshalIndent(proxies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, bytes, 0644)
}

func CheckProxy(p Proxy) bool {
	switch p.Type {
	case "http":
		return proxy.CheckOutproxySocksHTTP(p.Address, p.Port)
	case "https":
		return proxy.CheckOutproxySocksHTTPs(p.Address, p.Port)
	case "socks":
		return proxy.CheckOutproxySocksChain(p.Address, p.Port)
	default:
		log.Println("Unknown proxy type:", p.Type)
		return false
	}
}
