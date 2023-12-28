package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/bingoohuang/gocmd"
	"github.com/bingoohuang/gocmd/shellquote"
	"github.com/bingoohuang/gum/confirm"
	"github.com/imroc/req/v3"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

type FirewallCmd struct {
	InstanceId string `short:"i" help:"InstanceId."`
	File       string `short:"f" help:"防火墙规则JSON文件, 请查询后修改, e.g. firewall-xxx.json"`
}

type InstanceFirewallRules struct {
	InstanceId *string
	Rules      []lighthouse.FirewallRule
}

func (r *FirewallCmd) Run(*Context) error {
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

	type PingResult struct {
		RemoteAddr   string `json:"RemoteAddr"`
		ForwardedFor string `json:"X-Forwarded-For"`
		RealIP       string `json:"X-Real-IP"`
	}

	var pingResult PingResult
	reqClient.R().SetSuccessResult(&pingResult).Get("https://d5k.top/ping")
	pingResultJSON, _ := json.Marshal(pingResult)
	log.Printf("ping: %s", pingResultJSON)

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

	log.Printf("cmd: %s firewall -f %q", os.Args[0], file.Name())

	c := gocmd.New(shellquote.QuoteMust("code", file.Name()))
	if err = c.Run(context.Background()); err != nil {
		return nil
	}

	confirmOptions := &confirm.Options{}
	confirm, err := confirmOptions.Confirm("确认修改防火墙规则么?")
	if err != nil {
		return err
	}

	if confirm != "YES" {
		return nil
	}

	if err := r.modifyRules(file.Name()); err != nil {
		return err
	}
	return os.Remove(file.Name())
}

var reqClient = req.C()

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
