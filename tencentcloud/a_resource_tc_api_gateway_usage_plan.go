/*
Use this resource to create api gateway usage plan.

Example Usage

```hcl
resource "tencentcloud_api_gateway_usage_plan" "plan" {
  usage_plan_name         = "my_plan"
  usage_plan_desc         = "nice plan"
  max_request_num         = 100
  max_request_num_pre_sec = 10
}
```

Import

api gateway usage plan can be imported using the id, e.g.

```
$ terraform import tencentcloud_api_gateway_usage_plan.plan usagePlan-gyeafpab
```

*/
package tencentcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudAPIGatewayUsagePlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayUsagePlanCreate,
		Read:   resourceTencentCloudAPIGatewayUsagePlanRead,
		Update: resourceTencentCloudAPIGatewayUsagePlanUpdate,
		Delete: resourceTencentCloudAPIGatewayUsagePlanDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"usage_plan_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Custom usage plan name.",
			},
			"usage_plan_desc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom usage plan description.",
			},
			"max_request_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					if value := int64(i.(int)); value == -1 {
						return
					}
					return validateIntegerInRange(1, 99999999)(i, s)
				},
				Description: "Total number of requests allowed. Valid values: -1, [1,99999999]. The default value is -1, which indicates no limit.",
			},
			"max_request_num_pre_sec": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					if value := int64(i.(int)); value == -1 {
						return
					}
					return validateIntegerInRange(1, 2000)(i, s)
				},
				Description: "Limit of requests per second. Valid values: -1, [1,2000]. The default value is -1, which indicates no limit.",
			},

			// Computed values.
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
			"attach_api_keys": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "Attach api keys list.",
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
	}
}

func resourceTencentCloudAPIGatewayUsagePlanCreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan.create")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		usagePlanName       = data.Get("usage_plan_name").(string)
		usagePlanDesc       *string
		maxRequestNum       = int64(data.Get("max_request_num").(int))
		maxRequestNumPreSec = int64(data.Get("max_request_num_pre_sec").(int))

		usagePlanId   string
		inErr, outErr error
	)

	if v, has := data.GetOk("usage_plan_desc"); has {
		usagePlanDesc = helper.String(v.(string))
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		usagePlanId, inErr = apiGatewayService.CreateUsagePlan(ctx,
			usagePlanName,
			usagePlanDesc,
			maxRequestNum,
			maxRequestNumPreSec)

		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}

	data.SetId(usagePlanId)

	//wait usage plan create ok
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr := apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr)
		}
		if has {
			return nil
		}
		return resource.RetryableError(fmt.Errorf(" usage plan  %s not found on server", usagePlanId))

	}); outErr != nil {
		return outErr
	}

	return resourceTencentCloudAPIGatewayUsagePlanRead(data, meta)

}
func resourceTencentCloudAPIGatewayUsagePlanRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan.create")()
	defer inconsistentCheck(data, meta)()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		usagePlanId = data.Id()

		inErr      error
		info       apigateway.UsagePlanInfo
		has        bool
		attachList []*apigateway.UsagePlanEnvironment
	)

	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	if !has {
		data.SetId("")
		return nil
	}

	//service attach and api
	for _, bindType := range API_GATEWAY_TYPES {
		if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
			list, inErr := apiGatewayService.DescribeUsagePlanEnvironments(ctx, usagePlanId, bindType)
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

	errs := []error{
		data.Set("usage_plan_name", info.UsagePlanName),
		data.Set("usage_plan_desc", info.UsagePlanDesc),
		data.Set("max_request_num", info.MaxRequestNum),
		data.Set("max_request_num_pre_sec", info.MaxRequestNumPreSec),
		data.Set("modify_time", info.ModifiedTime),
		data.Set("create_time", info.CreatedTime),
		data.Set("attach_list", infoAttachList),
	}

	attach_api_keys := make([]string, 0, len(info.BindSecretIds))
	for _, v := range info.BindSecretIds {
		attach_api_keys = append(attach_api_keys, *v)
	}
	errs = append(errs, data.Set("attach_api_keys", attach_api_keys))

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
func resourceTencentCloudAPIGatewayUsagePlanUpdate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan.update")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		usagePlanName       = data.Get("usage_plan_name").(string)
		usagePlanDesc       = helper.String(data.Get("usage_plan_desc").(string))
		maxRequestNum       = int64(data.Get("max_request_num").(int))
		maxRequestNumPreSec = int64(data.Get("max_request_num_pre_sec").(int))

		usagePlanId   = data.Id()
		inErr, outErr error
	)

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr = apiGatewayService.ModifyUsagePlan(ctx,
			usagePlanId,
			usagePlanName,
			usagePlanDesc,
			maxRequestNum,
			maxRequestNumPreSec)
		_ = inErr
		//
		//if inErr != nil {
		//	return retryError(inErr)
		//}
		return nil
	})
	if outErr != nil {
		return outErr
	}

	return resourceTencentCloudAPIGatewayUsagePlanRead(data, meta)
}
func resourceTencentCloudAPIGatewayUsagePlanDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan.delete")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		usagePlanId       = data.Id()
	)

	return resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr := apiGatewayService.DeleteUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
}
