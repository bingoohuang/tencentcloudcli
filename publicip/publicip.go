package publicip

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

var Endpoints = []string{
	"https://api.maao.cc/ip/",
	"https://d5k.top/ping",
	"https://api.ipify.org?format=json",
	"https://httpbin.org/ip",
	"ip.gs",
	"ip.sb",
	"cip.cc",
	"icanhazip.com",
	"api.ipify.org",
	"ipinfo.io/ip",
	"ifconfig.me",
	"ifconfig.co",
	"ipecho.net/plain",
	"whatismyip.akamai.com",
	"inet-ip.info",
	"myip.ipip.net",
}

func init() {
	for i, ipUrl := range Endpoints {
		if !strings.HasPrefix(ipUrl, "http") {
			Endpoints[i] = "http://" + ipUrl
		}
	}
}

func CheckPublicIP() {
	var wg sync.WaitGroup
	for _, ipUrl := range Endpoints {
		wg.Add(1)
		go func(ipUrl string) {
			defer wg.Done()
			invoke(ipUrl)
		}(ipUrl)
	}

	wg.Wait()
}

var cutBlanks = regexp.MustCompile(`\s+`)

func invoke(ipUrl string) {
	if res, err := client.R().
		SetHeader("User-Agent", "curl").
		Get(ipUrl); err == nil {
		if data := res.Bytes(); len(data) > 0 {
			data := cutBlanks.ReplaceAll(data, []byte(" "))
			log.Printf("%s: %s", ipUrl, data)
		}
	}
}

var client = req.C().
	SetTimeout(15 * time.Second).
	SetProxy(nil) // Disable proxy
