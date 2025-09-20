package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"os"
	"outproxys/proxy"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

type Proxy struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Uptime  int    `json:"uptime"`
	Type    string `json:"type"` // "http", "https", "socks"
}

var (
	dataFile = "proxies.json"
	mu       sync.Mutex
	reB32    = regexp.MustCompile(`^[a-z2-7]{52}\.b32\.i2p$`)
	reI2P    = regexp.MustCompile(`^[a-z0-9\-\.]+\.i2p$`)
)

func loadProxies() ([]Proxy, error) {
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

func saveProxies(proxies []Proxy) error {
	bytes, err := json.MarshalIndent(proxies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, bytes, 0644)
}

func checkProxy(p Proxy) bool {
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

func monitorProxies() {
	for {
		time.Sleep(30 * time.Second)

		mu.Lock()
		proxies, err := loadProxies()
		if err != nil {
			mu.Unlock()
			continue
		}
		changed := false
		for i := range proxies {
			ok := checkProxy(proxies[i])
			if ok && proxies[i].Uptime < 100 {
				proxies[i].Uptime++
				changed = true
			}
			if !ok && proxies[i].Uptime > 0 {
				proxies[i].Uptime--
				changed = true
			}
		}
		if changed {
			saveProxies(proxies)
		}
		mu.Unlock()
	}
}

type PageData struct {
	Proxies  []Proxy
	PrevLink string
	NextLink string
}

func main() {
	loadConfig()
	log.Println("Run server")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("public/index.html"))

		mu.Lock()
		proxies, _ := loadProxies()
		mu.Unlock()

		// сортировка по Uptime
		sort.Slice(proxies, func(i, j int) bool {
			return proxies[i].Uptime > proxies[j].Uptime
		})

		perPage := 20
		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
			page = p
		}

		start := (page - 1) * perPage
		end := start + perPage
		if start > len(proxies) {
			start = len(proxies)
		}
		if end > len(proxies) {
			end = len(proxies)
		}

		proxiesPage := proxies[start:end]

		prevLink := ""
		nextLink := ""
		if start > 0 {
			prevLink = fmt.Sprintf("/?page=%d", page-1)
		}
		if end < len(proxies) {
			nextLink = fmt.Sprintf("/?page=%d", page+1)
		}

		pd := PageData{
			Proxies:  proxiesPage,
			PrevLink: prevLink,
			NextLink: nextLink,
		}

		err := tmpl.Execute(w, pd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		address := r.FormValue("address")
		portStr := r.FormValue("port")
		typeStr := r.FormValue("type")
		if typeStr != "http" && typeStr != "https" && typeStr != "socks" {
			http.Error(w, "invalid type proxy", http.StatusBadRequest)
			return
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			http.Error(w, "invalid port", http.StatusBadRequest)
			return
		}

		if !reB32.MatchString(address) && !reI2P.MatchString(address) {
			http.Error(w, "invalid i2p address", http.StatusBadRequest)
			return
		}

		if !checkProxy(Proxy{Address: address, Port: port, Type: typeStr, Uptime: 50}) {
			http.Error(w, "not a valid outproxy", http.StatusBadRequest)
			return
		}

		mu.Lock()
		proxies, _ := loadProxies()
		proxies = append(proxies, Proxy{Address: address, Port: port, Type: typeStr, Uptime: 50})
		saveProxies(proxies)
		mu.Unlock()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))

	go monitorProxies()
	host := viper.GetString("server.host")
	port := strconv.Itoa(viper.GetInt("server.port"))

	s := &http.Server{
		Addr:           host + ":" + port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Listen on host " + host)
	log.Println("Listen server on port " + port)
	log.Fatal(s.ListenAndServe())
}
