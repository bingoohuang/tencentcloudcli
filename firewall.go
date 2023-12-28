package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/gum/confirm"
	"github.com/commander-cli/cmd"
	"github.com/imroc/req/v3"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

type FirewallCmd struct {
	InstanceId string `short:"i" help:"InstanceId."`
	File       string `short:"f" help:"防火墙规则JSON文件, 请查询后修改, e.g. firewall-xxx.json"`
	PublicIP   bool   `short:"p" help:"查询公网出口IP"`
}

type InstanceFirewallRules struct {
	InstanceId *string
	Rules      []lighthouse.FirewallRule
}

func (r *FirewallCmd) Run(*Context) error {
	if r.PublicIP {
		checkPublicIP()
		return nil
	}

	if r.File != "" {
		return r.modifyRules(r.File)
	}
	return r.listRules()
}

func (r *FirewallCmd) listRules() error {
	req := lighthouse.NewDescribeFirewallRulesRequest()
	if r.InstanceId == "" {
		if parts := getSecretParts(); len(parts) > 2 {
			r.InstanceId = parts[2]
		}
	}
	req.InstanceId = &r.InstanceId

	// https://console.cloud.tencent.com/api/explorer?Product=lighthouse&Version=2020-03-24&Action=DescribeFirewallRules
	// 返回的resp是一个DescribeFirewallRulesResponse的实例，与请求对象对应
	response, err := getClient().DescribeFirewallRules(req)
	if err != nil {
		return err
	}

	rules := InstanceFirewallRules{
		InstanceId: req.InstanceId,
	}
	for _, rule := range response.Response.FirewallRuleSet {
		rules.Rules = append(rules.Rules, lighthouse.FirewallRule{
			Protocol:                rule.Protocol,
			Port:                    rule.Port,
			CidrBlock:               rule.CidrBlock,
			Action:                  rule.Action,
			FirewallRuleDescription: rule.FirewallRuleDescription,
		})
	}

	checkPublicIP()

	jsonRules, err := json.MarshalIndent(rules, "", "    ")
	if err != nil {
		return err
	}

	// Create a temporary file
	file, errs := os.CreateTemp("", "temp-*.txt")
	if errs != nil {
		return err
	}

	defer file.Close()

	if _, err := file.Write(jsonRules); err != nil {
		return err
	}

	file.Close()

	log.Printf("请修改防火墙规则: %s", file.Name())

	c := cmd.NewCommand("code " + file.Name())
	if err = c.Execute(); err == nil {
		confirmOptions := &confirm.Options{}
		confirm, err := confirmOptions.Confirm("确认修改防火墙规则么?")
		if err != nil {
			return err
		}

		if confirm == "YES" {
			if err := r.modifyRules(file.Name()); err != nil {
				return err
			}
			os.Remove(file.Name())
			return err
		}
	}

	fmt.Printf("单独执行 `tencentcloudcli firewall --file=%s` 完成防火墙规则修改!\n", file.Name())
	return nil
}

func checkPublicIP() {
	var wg sync.WaitGroup
	for _, ipUrl := range []string{
		"https://d5k.top/ping",
		"https://api.ipify.org?format=json",
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
	} {
		wg.Add(1)
		go func(ipUrl string) {
			defer wg.Done()
			if !strings.HasPrefix(ipUrl, "http") {
				ipUrl = "http://" + ipUrl
			}
			if res, _ := reqClient.R().SetHeader("User-Agent", "curl").Get(ipUrl); res != nil {
				if data := res.Bytes(); len(data) > 0 {
					re := regexp.MustCompile(`\s+`)
					data := re.ReplaceAll(data, []byte(" "))
					log.Printf("%s: %s", ipUrl, data)
				}
			}
		}(ipUrl)
	}

	wg.Wait()
}

var reqClient = req.C().SetTimeout(15 * time.Second)

func (r *FirewallCmd) modifyRules(file string) error {
	if _, err := os.Stat(file); err != nil {
		return err
	}

	fileData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var rules InstanceFirewallRules
	if err := json.Unmarshal(fileData, &rules); err != nil {
		return err
	}

	rq := lighthouse.NewModifyFirewallRulesRequest()
	rq.InstanceId = rules.InstanceId
	for _, rule := range rules.Rules {
		rq.FirewallRules = append(rq.FirewallRules, &lighthouse.FirewallRule{
			Protocol:                rule.Protocol,
			Port:                    rule.Port,
			CidrBlock:               rule.CidrBlock,
			Action:                  rule.Action,
			FirewallRuleDescription: rule.FirewallRuleDescription,
		})
	}

	rsp, err := getClient().ModifyFirewallRules(rq)
	if err != nil {
		return err
	}

	resp, _ := json.Marshal(rsp)
	log.Printf("ModifyFirewallRules: %s", resp)
	return nil
}
