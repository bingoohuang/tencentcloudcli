package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

type FirewallCmd struct {
	InstanceId string `help:"InstanceId."`
	File       string `help:"防火桥规则JSON文件，请查询后修改, e.g. firewall-xxx.json"`
}

type InstanceFirewallRules struct {
	InstanceId *string
	Rules      []lighthouse.FirewallRule
}

func (r *FirewallCmd) Run(*Context) error {
	if r.File != "" {
		return r.modifyRules()
	}
	return r.listRules()
}

func (r *FirewallCmd) modifyRules() error {
	if _, err := os.Stat(r.File); err != nil {
		return err
	}

	fileData, err := os.ReadFile(r.File)
	if err != nil {
		return err
	}

	var rules InstanceFirewallRules
	if err := json.Unmarshal(fileData, &rules); err != nil {
		return err
	}

	req := lighthouse.NewModifyFirewallRulesRequest()
	req.InstanceId = rules.InstanceId
	for _, rule := range rules.Rules {
		req.FirewallRules = append(req.FirewallRules, &lighthouse.FirewallRule{
			Protocol:                rule.Protocol,
			Port:                    rule.Port,
			CidrBlock:               rule.CidrBlock,
			Action:                  rule.Action,
			FirewallRuleDescription: rule.FirewallRuleDescription,
		})
	}

	rsp, err := getClient().ModifyFirewallRules(req)
	if err != nil {
		return err
	}

	resp, _ := json.MarshalIndent(rsp, "", "    ")
	fmt.Printf("%s\n", resp)
	return nil
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

	jsonRules, err := json.MarshalIndent(rules, "", "    ")
	if err != nil {
		return err
	}

	jsonFile := "firewall_rules_" + time.Now().Format(`20060102150405`) + ".json"
	if err := os.WriteFile(jsonFile, jsonRules, os.ModePerm); err != nil {
		fmt.Printf("%s\n", jsonFile)
	} else {
		fmt.Printf("%s 已生成，请修改此文件后执行 `tencentcloudcli firewall --file=%s` 完成防火墙规则修改!\n", jsonFile, jsonFile)
	}

	return nil
}
