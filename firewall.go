package main

import (
	"context"
	"encoding/json"
	"github.com/bingoohuang/gocmd"
	"github.com/bingoohuang/gocmd/shellquote"
	"github.com/bingoohuang/gum/confirm"
	"github.com/bingoohuang/tencentcloudcli/publicip"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"log"
	"os"
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

func (r *FirewallCmd) Run(_ *Context) error {
	if r.PublicIP {
		publicip.CheckPublicIP()
		return nil
	}

	if r.File != "" {
		return r.modifyRules(r.File)
	}
	return r.listRules()
}

func (r *FirewallCmd) listRules() error {
	rq := lighthouse.NewDescribeFirewallRulesRequest()
	if r.InstanceId == "" {
		r.InstanceId = LightHouse.InstanceId
	}
	rq.InstanceId = &r.InstanceId

	// https://console.cloud.tencent.com/api/explorer?Product=lighthouse&Version=2020-03-24&Action=DescribeFirewallRules
	// 返回的resp是一个DescribeFirewallRulesResponse的实例，与请求对象对应
	response, err := getClient().DescribeFirewallRules(rq)
	if err != nil {
		return err
	}

	rules := InstanceFirewallRules{
		InstanceId: rq.InstanceId,
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

	go publicip.CheckPublicIP()

	jsonRules, err := json.MarshalIndent(rules, "", "    ")
	if err != nil {
		return err
	}

	// Create a temporary file
	file, err := TempFile(jsonRules)
	if err != nil {
		return err
	}

	log.Printf("cmd: %s", shellquote.QuoteMust(os.Args[0], "firewall", "-f", file))

	c := gocmd.New(shellquote.QuoteMust("code", file))
	if err = c.Run(context.Background()); err != nil {
		return nil
	}

	confirmOptions := &confirm.Options{}
	yes, err := confirmOptions.Confirm("确认修改防火墙规则么?")
	if err != nil {
		return err
	}

	if yes != "YES" {
		return nil
	}

	if err := r.modifyRules(file); err != nil {
		return err
	}
	return os.Remove(file)
}

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

// TempFile 创建临时文件，写入内容 data
func TempFile(data []byte) (string, error) {
	f, err := os.CreateTemp("", "*")
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}
