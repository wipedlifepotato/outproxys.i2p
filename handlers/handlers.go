package handlers

import "html/template"
import "fmt"
import "net/http"
import "sort"
import "regexp"
import "github.com/spf13/viper"
import "strconv"
import (
	"outproxys/goroutines"
	"outproxys/models"
)

var (
	reB32 = regexp.MustCompile(`^[a-z2-7]{52}\.b32\.i2p$`)
	reI2P = regexp.MustCompile(`^[a-z0-9\-\.]+\.i2p$`)
)

func AddHandle(w http.ResponseWriter, r *http.Request) {
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
	newProxy := models.Proxy{
		Address: address,
		Port:    port,
		Type:    typeStr,
		Uptime:  50,
	}
	if !models.CheckProxy(newProxy) {
		http.Error(w, "not a valid outproxy", http.StatusBadRequest)
		return
	}

	goroutines.Mu.Lock()
	proxies, _ := models.LoadProxies()
	for _, p := range proxies {
		if p.Address == newProxy.Address && p.Port == newProxy.Port {
			http.Error(w, "proxy already exists", http.StatusConflict)
			return
		}
	}
	proxies = append(proxies, models.Proxy{Address: address, Port: port, Type: typeStr, Uptime: 50})
	models.SaveProxies(proxies)
	goroutines.Mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func MainHandle(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/index.html"))

	goroutines.Mu.Lock()
	proxies, err := models.LoadProxies()
	if err != nil {
		http.Error(w, "Can't read proxies", http.StatusBadRequest)
		return
	}
	goroutines.Mu.Unlock()

	sort.Slice(proxies, func(i, j int) bool {
		return proxies[i].Uptime > proxies[j].Uptime
	})

	perPage := viper.GetInt("page.proxyPerPage")
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

	pd := models.PageData{
		Proxies:  proxiesPage,
		PrevLink: prevLink,
		NextLink: nextLink,
		Donation: viper.GetString("page.donation"),
	}

	err = tmpl.Execute(w, pd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
