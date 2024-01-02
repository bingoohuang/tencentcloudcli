# tencentcloudcli 腾讯云工具

## 轻量服务器防火墙规则

```shell
# 设置secretId、secretKey，及轻量级服务器 instanceId
$ export LIGHTHOUSE_SECRET=secretId:secretKey:instanceId

# 查询当前防火墙规则
$ tencentcloudcli firewall                                                                                   
firewall_rules_20231108145925.json 已生成，请修改此文件后执行 `tencentcloudcli firewall --file=firewall_rules_20231108145925.json` 完成防火墙规则修改!

# 编辑防火墙规则 JSON 文件，然后执行防火墙规则修改
$ tencentcloudcli firewall --file=firewall_rules_20231108145925.json
{
    "Response": {
        "RequestId": "58496587-4bcf-47e7-9e78-1303eda4cb54"
    }
}
```
