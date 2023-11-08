package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

var (
	_clientOnce sync.Once
	_client     *lighthouse.Client
)

func getClient() *lighthouse.Client {
	_clientOnce.Do(func() {
		// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
		// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，
		// 建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
		// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
		parts := getSecretParts()
		c := common.NewCredential(parts[0], parts[1])
		// 实例化一个client选项，可选的，没有特殊需求可以跳过
		p := profile.NewClientProfile()
		p.HttpProfile.Endpoint = getEnv("LIGHTHOUSE_ENDPOINT", "lighthouse.tencentcloudapi.com")
		// 实例化要请求产品的client对象,clientProfile是可选的
		var err error
		_client, err = lighthouse.NewClient(c, getEnv("REGION", "ap-beijing"), p)
		if err != nil {
			panic(err)
		}
	})

	return _client
}

func getEnv(name, defaultValue string) string {
	if env := os.Getenv(name); env != "" {
		return env
	}

	return defaultValue
}

func getSecretParts() []string {
	secret := os.Getenv("LIGHTHOUSE_SECRET")
	if secret == "" {
		log.Fatalf("please set env, e.g. export LIGHTHOUSE_SECRET=secretId:secretKey")
	}
	parts := strings.Split(secret, ":")
	if len(parts) < 2 {
		log.Fatalf("bad $LIGHTHOUSE_SECRET")
	}
	return parts
}
