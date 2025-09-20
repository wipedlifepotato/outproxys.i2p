package goroutines

import "time"
import "outproxys/models"
import "github.com/spf13/viper"
import "sync"

var Mu sync.Mutex

func MonitorProxies() {
	for {
		time.Sleep(time.Duration(viper.GetInt("monitor.interval_seconds")) * time.Second)

		Mu.Lock()
		proxies, err := models.LoadProxies()
		if err != nil {
			Mu.Unlock()
			continue
		}
		changed := false
		for i := range proxies {
			ok := models.CheckProxy(proxies[i])
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
			models.SaveProxies(proxies)
		}
		Mu.Unlock()
	}
}
