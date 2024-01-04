package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bingoohuang/tencentcloudcli/tmpjson"
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
		c := common.NewCredential(LightHouse.SecretID, LightHouse.SecretKey)
		// 实例化一个client选项，可选的，没有特殊需求可以跳过
		p := profile.NewClientProfile()
		p.HttpProfile.Endpoint = LightHouse.Endpoint
		// 实例化要请求产品的client对象,clientProfile是可选的
		var err error
		_client, err = lighthouse.NewClient(c, LightHouse.Region, p)
		if err != nil {
			panic(err)
		}
	})

	return _client
}

const jsonFile = "lighthouse.json"

var LightHouse = func() (lh LightHouseConf) {
	jsonFileRewrite := false

	if env := os.Getenv("LIGHTHOUSE_SECRET"); env != "" {
		parts := strings.Split(env, ":")
		if len(parts) < 2 {
			log.Printf("bad $LIGHTHOUSE_SECRET")
		} else {
			lh.SecretID = parts[0]
			lh.SecretKey = parts[1]
			if len(parts) > 2 {
				lh.InstanceId = parts[2]
			}
			jsonFileRewrite = true
		}
	}

	if lh.SecretID == "" || lh.SecretKey == "" || lh.InstanceId == "" || lh.Region == "" {
		if _, err := tmpjson.Read(jsonFile, &lh); err != nil {
			log.Fatalf("please set env, e.g. export LIGHTHOUSE_SECRET=secretId:secretKey")
		}
		jsonFileRewrite = false
	}

	if lh.Region == "" {
		lh.Region = "ap-beijing"
		jsonFileRewrite = true
	}
	if lh.Endpoint == "" {
		lh.Endpoint = "lighthouse.tencentcloudapi.com"
		jsonFileRewrite = true
	}

	if jsonFileRewrite {
		if err := tmpjson.Write(jsonFile, lh); err != nil {
			log.Printf("write %s error: %v", jsonFile, err)
		}
	}

	return lh
}()

type LightHouseConf struct {
	SecretID   string `json:"secretID"`
	SecretKey  string `json:"secretKey"`
	InstanceId string `json:"instanceId"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
}
