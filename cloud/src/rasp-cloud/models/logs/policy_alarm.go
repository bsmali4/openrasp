//Copyright 2017-2019 Baidu Inc.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http: //www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package logs

import (
	"fmt"
	"crypto/md5"
	"github.com/astaxie/beego"
	"time"
	"rasp-cloud/conf"
)

var (
	PolicyAlarmInfo = AlarmLogInfo{
		EsType:       "policy-alarm",
		EsIndex:      "openrasp-policy-alarm",
		EsAliasIndex: "real-openrasp-policy-alarm",
		TtlTime:      24 * 365 * time.Hour,
		AlarmBuffer:  make(chan map[string]interface{}, conf.AppConfig.AlarmBufferSize),
		FileLogger:   initAlarmFileLogger("/openrasp-logs/policy-alarm", "policy.log"),
		EsMapping: `
		{
			"settings": {
				"analysis": {
					"normalizer": {
						"lowercase_normalizer": {
							"type": "custom",
							"filter": ["lowercase"]
						}
					}
				}
			},
			"mappings": {
				"policy-alarm": {
					"_all": {
						"enabled": false
					},
					"properties": {
						"@timestamp":{
							"type":"date"
						},
						"event_type": {
							"type": "keyword",
							"ignore_above": 256
						},
						"server_hostname": {
							"type": "keyword",
							"ignore_above": 256,
							"normalizer": "lowercase_normalizer"
						},
						"server_type": {
							"type": "keyword",
							"ignore_above": 64
						},
						"server_nic": {
							"type": "nested",
							"properties": {
								"name": {
									"type": "keyword",
									"ignore_above": 256
								},
								"ip": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						},
						"app_id": {
							"type": "keyword",
							"ignore_above": 256
						},
						"rasp_id": {
							"type": "keyword",
							"ignore_above": 256
						},
						"event_time": {
							"type": "date"
						},
						"stack_trace": {
							"type": "keyword"
						},
						"policy_id": {
							"type": "long"
						},
						"message": {
							"type": "keyword"
						},
						"stack_md5": {
							"type": "keyword",
							"ignore_above": 64
						},
						"policy_params": {
							"type": "object",
							"enabled":"false"
						}
					}
				}
			}
		}
		`,
	}
)

func init() {
	registerAlarmInfo(&PolicyAlarmInfo)
}

func AddPolicyAlarm(alarm map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			beego.Error("failed to add policy alarm: ", r)
		}
	}()
	if stack, ok := alarm["stack_trace"]; ok && stack != nil && stack != "" {
		_, ok = stack.(string)
		if ok {
			alarm["stack_md5"] = fmt.Sprintf("%x", md5.Sum([]byte(stack.(string))))
		}
	}
	idContent := ""
	idContent += fmt.Sprint(alarm["rasp_id"])
	idContent += fmt.Sprint(alarm["policy_id"])
	idContent += fmt.Sprint(alarm["stack_md5"])
	if alarm["policy_id"] == "3006" && alarm["policy_params"] != nil {
		if policyParam, ok := alarm["policy_params"].(map[string]interface{}); ok && len(policyParam) > 0 {
			idContent += fmt.Sprint(policyParam["connectionString"])
			idContent += fmt.Sprint(policyParam["port"])
			idContent += fmt.Sprint(policyParam["server"])
			idContent += fmt.Sprint(policyParam["hostname"])
			idContent += fmt.Sprint(policyParam["socket"])
			idContent += fmt.Sprint(policyParam["username"])
		}
	}
	alarm["upsert_id"] = fmt.Sprintf("%x", md5.Sum([]byte(idContent)))
	return AddAlarmFunc(PolicyAlarmInfo.EsType, alarm)
}
