package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"os"
	"golang.org/x/net/proxy"
)

var (
	addrToCheck     = "https://check.torproject.org"
	localSocksProxy = getEnv("I2PD_SOCKS_HOST", "127.0.0.1") + ":" + getEnv("I2PD_SOCKS_PORT", "4447")
)

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func CheckOutproxySocksHTTPs(address string, port int) bool {

	socksAddr := localSocksProxy

	targetProxy := fmt.Sprintf("http://%s:%d", address, port)
	proxyURL, err := url.Parse(targetProxy)
	if err != nil {
		log.Println("Invalid proxy URL:", err)
		return false
	}

	socksDialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("SOCKS5 dialer error:", err)
		return false
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		Dial:  socksDialer.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}

	resp, err := client.Get(addrToCheck)
	if err != nil {
		log.Println("GET error:", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("Response status:", resp.Status)
	log.Println("Body first 200 chars:\n", string(body)[:200])

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

//

func CheckOutproxySocksChain(address string, port int) bool {
	localSocks := localSocksProxy
	targetSocks := fmt.Sprintf("%s:%d", address, port)

	innerDialer, err := proxy.SOCKS5("tcp", localSocks, nil, proxy.Direct)
	if err != nil {
		log.Println("Target SOCKS5 dialer error:", err)
		return false
	}

	outerDialer, err := proxy.SOCKS5("tcp", targetSocks, nil, innerDialer)
	if err != nil {
		log.Println("Local SOCKS5 dialer error:", err)
		return false
	}

	transport := &http.Transport{
		Dial: outerDialer.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // для HTTPS через SOCKS
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}

	resp, err := client.Get(addrToCheck)
	if err != nil {
		log.Println("GET error:", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("Response status:", resp.Status)
	log.Println("Body first 200 chars:\n", string(body)[:200])

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

///

func CheckOutproxySocksHTTP(address string, port int) bool {
	socksAddr := localSocksProxy
	targetProxy := fmt.Sprintf("http://%s:%d", address, port)

	proxyURL, err := url.Parse(targetProxy)
	if err != nil {
		log.Println("Invalid proxy URL:", err)
		return false
	}

	// SOCKS5 dialer
	socksDialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("SOCKS5 dialer error:", err)
		return false
	}

	// HTTP-транспорт через SOCKS5
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		Dial:  socksDialer.Dial, // Dial для TCP через SOCKS5
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	// HTTP-запрос
	resp, err := client.Get("http://check.torproject.org") // <-- HTTP
	if err != nil {
		log.Println("GET error:", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("Response status:", resp.Status)
	log.Println("Body first 200 chars:\n", string(body)[:200])

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
