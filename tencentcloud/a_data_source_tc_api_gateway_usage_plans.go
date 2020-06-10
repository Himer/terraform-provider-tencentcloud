/*
Use this data source to query api gateway usage plans.

Example Usage

```hcl
resource "tencentcloud_api_gateway_usage_plan" "plan" {
  usage_plan_name         = "my_plan"
  usage_plan_desc         = "nice plan"
  max_request_num         = 100
  max_request_num_pre_sec = 10
}

data "tencentcloud_api_gateway_usage_plans" "name" {
  usage_plan_name = tencentcloud_api_gateway_usage_plan.plan.usage_plan_name
}

data "tencentcloud_api_gateway_usage_plans" "id" {
  usage_plan_id = tencentcloud_api_gateway_usage_plan.plan.id
}
```
*/
package tencentcloud

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
	"log"
)

func dataSourceTencentCloudAPIGatewayUsagePlans() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudAPIGatewayUsagePlansRead,

		Schema: map[string]*schema.Schema{
			"usage_plan_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the usage plan to be queried.",
			},
			"usage_plan_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the usage plan to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},

			// Computed values.
			"list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of usage plans. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"usage_plan_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the usage plan.",
						},
						"usage_plan_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the usage plan.",
						},
						"usage_plan_desc": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Custom usage plan description.",
						},
						"max_request_num": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total number of requests allowed. Valid values: -1, [1,99999999]. The default value is -1, which indicates no limit.",
						},
						"max_request_num_pre_sec": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Limit of requests per second. Valid values: -1, [1,2000]. The default value is -1, which indicates no limit.",
						},
						"modify_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last modified time in the format of YYYY-MM-DDThh:mm:ssZ according to ISO 8601 standard. UTC time is used.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time in the format of YYYY-MM-DDThh:mm:ssZ according to ISO 8601 standard. UTC time is used.",
						},
						"attach_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Attach service and api list.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The service id.",
									},
									"service_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The service name.",
									},
									"api_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api id, This value is empty if attach service.",
									},
									"api_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api name, This value is empty if attach service.",
									},
									"path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api path, This value is empty if attach service.",
									},
									"method": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api method, This value is empty if attach service.",
									},
									"environment": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The environment name.",
									},
									"modify_time": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Last modified time in the format of YYYY-MM-DDThh:mm:ssZ according to ISO 8601 standard. UTC time is used.",
									},
									"create_time": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Creation time in the format of YYYY-MM-DDThh:mm:ssZ according to ISO 8601 standard. UTC time is used.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceTencentCloudAPIGatewayUsagePlansRead(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_api_gateway_usage_plans.read")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		usagePlanId   = data.Get("usage_plan_id").(string)
		usagePlanName = data.Get("usage_plan_name").(string)
		infos         []*apigateway.UsagePlanStatusInfo

		inErr error
	)

	if outErr := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		infos, inErr = apiGatewayService.DescribeUsagePlansStatus(ctx, usagePlanId, usagePlanName)
		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	var list []map[string]interface{}

	for _, info := range infos {

		var infoMap = make(map[string]interface{}, 10)
		var attachList []*apigateway.UsagePlanEnvironment

		for _, bindType := range API_GATEWAY_TYPES {
			if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
				list, inErr := apiGatewayService.DescribeUsagePlanEnvironments(ctx, *info.UsagePlanId, bindType)
				if inErr != nil {
					return retryError(inErr)
				}
				attachList = append(attachList, list...)
				return nil
			}); outErr != nil {
				return outErr
			}
		}

		infoAttachList := make([]map[string]interface{}, 0, len(attachList))
		for _, v := range attachList {
			infoAttachList = append(infoAttachList, map[string]interface{}{
				"service_id":   v.ServiceId,
				"service_name": v.ServiceName,
				"api_id":       v.ApiId,
				"api_name":     v.ApiName,
				"path":         v.Path,
				"method":       v.Method,
				"environment":  v.Environment,
				"modify_time":  v.ModifiedTime,
				"create_time":  v.CreatedTime,
			})
		}
		infoMap["usage_plan_id"] = info.UsagePlanId
		infoMap["usage_plan_name"] = info.UsagePlanName
		infoMap["usage_plan_desc"] = info.UsagePlanDesc
		infoMap["max_request_num"] = info.MaxRequestNum
		infoMap["max_request_num_pre_sec"] = info.MaxRequestNumPreSec
		infoMap["modify_time"] = info.ModifiedTime
		infoMap["create_time"] = info.CreatedTime
		infoMap["attach_list"] = infoAttachList

		list = append(list, infoMap)
	}

	byteId, err := json.Marshal(map[string]interface{}{
		"usage_plan_id":   usagePlanId,
		"usage_plan_name": usagePlanName,
	})
	if err != nil {
		return err
	}

	if err = data.Set("list", list); err != nil {
		log.Printf("[CRITAL]%s provider set list fail, reason:%s\n ", logId, err.Error())
	}

	data.SetId(string(byteId))

	if output, ok := data.GetOk("result_output_file"); ok && output.(string) != "" {
		return writeToFile(output.(string), list)
	}
	return nil
}
