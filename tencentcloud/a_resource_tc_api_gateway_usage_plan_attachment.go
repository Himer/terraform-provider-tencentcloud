/*
Use this resource to attach api gateway usage plan to service.

Example Usage

```hcl
resource "tencentcloud_api_gateway_usage_plan" "plan" {
  usage_plan_name         = "my_plan"
  usage_plan_desc         = "nice plan"
  max_request_num         = 100
  max_request_num_pre_sec = 10
}
resource "tencentcloud_api_gateway_service" "service" {
  service_name = "niceservice"
  protocol     = "http&https"
  service_desc = "your nice service"
  net_type     = ["INNER", "OUTER"]
  ip_version   = "IPv4"
}

resource "tencentcloud_api_gateway_usage_plan_attachment" "attach" {
  usage_plan_id = tencentcloud_api_gateway_usage_plan.plan.id
  service_id    = tencentcloud_api_gateway_service.service.id
  environment   = "test"
  bind_type     = "SERVICE"
}
```

Import

api gateway usage plan attachment can be imported using the id, e.g.

```
$ terraform import tencentcloud_api_gateway_usage_plan_attachment.attach '{"api_id":"","bind_type":"SERVICE","environment":"test","service_id":"service-pkegyqmc","usage_plan_id":"usagePlan-26t0l0w3"}]'
```

*/
package tencentcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
)

func resourceTencentCloudAPIGatewayUsagePlanAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayUsagePlanAttachmentCreate,
		Read:   resourceTencentCloudAPIGatewayUsagePlanAttachmentRead,
		Delete: resourceTencentCloudAPIGatewayUsagePlanAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"usage_plan_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the usage plan.",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the service.",
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_SERVICE_ENVS),
				Description:  "Environment to be bound `test`,`prepub` or `release`.",
			},
			"bind_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      API_GATEWAY_TYPE_SERVICE,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_TYPES),
				Description:  "Binding type. Valid values: `API`, `SERVICE` (default value).",
			},
			"api_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "API id. This parameter will be required when `bind_type` is `API`.",
			},
		},
	}
}

func resourceTencentCloudAPIGatewayUsagePlanAttachmentCreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan_attachment.create")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		usagePlanId = data.Get("usage_plan_id").(string)
		serviceId   = data.Get("service_id").(string)
		environment = data.Get("environment").(string)
		bindType    = data.Get("bind_type").(string)
		apiId       = data.Get("api_id").(string)

		inErr, outErr error
		has           bool
	)

	if bindType == API_GATEWAY_TYPE_API && apiId == "" {
		return fmt.Errorf("parameter `api_ids` is required when `bind_type` is `API`")
	}

	//check usage plan
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	if !has {
		return fmt.Errorf("usage plan %s not exist", usagePlanId)
	}

	//check service
	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeService(ctx, serviceId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}
	if !has {
		return fmt.Errorf("service %s not exist", serviceId)
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr = apiGatewayService.BindEnvironment(ctx,
			serviceId,
			usagePlanId,
			environment,
			bindType,
			apiId)

		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	idMap, outErr := json.Marshal(map[string]interface{}{
		"usage_plan_id": usagePlanId,
		"service_id":    serviceId,
		"environment":   environment,
		"bind_type":     bindType,
		"api_id":        apiId})
	if outErr != nil {
		return fmt.Errorf("build id json fail,%s", outErr.Error())
	}

	data.SetId(string(idMap))

	return resourceTencentCloudAPIGatewayUsagePlanAttachmentRead(data, meta)

}
func resourceTencentCloudAPIGatewayUsagePlanAttachmentRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan_attachment.read")()
	defer inconsistentCheck(data, meta)()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		idMap = make(map[string]string)

		outErr, inErr error
		has           bool
	)

	if outErr = json.Unmarshal([]byte(data.Id()), &idMap); outErr != nil {
		return fmt.Errorf("id is broken,%s", outErr.Error())
	}

	var (
		usagePlanId = idMap["usage_plan_id"]
		serviceId   = idMap["service_id"]
		environment = idMap["environment"]
		bindType    = idMap["bind_type"]
		apiId       = idMap["api_id"]
	)
	if usagePlanId == "" || serviceId == "" || environment == "" || bindType == "" {
		return fmt.Errorf("id is broken")
	}
	if bindType == API_GATEWAY_TYPE_API && apiId == "" {
		return fmt.Errorf("id is broken")
	}

	//check usage plan
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	if !has {
		data.SetId("")
		return nil
	}

	//check service
	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeService(ctx, serviceId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}
	if !has {
		data.SetId("")
		return nil
	}

	var plans []*apigateway.ApiUsagePlan

	if bindType == API_GATEWAY_TYPE_API {
		if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
			plans, inErr = apiGatewayService.DescribeApiUsagePlan(ctx, serviceId)
			if inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
	} else {
		if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
			plans, inErr = apiGatewayService.DescribeServiceUsagePlan(ctx, serviceId)
			if inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
	}
	var setData = func() error {
		for _, err := range []error{data.Set("usage_plan_id", usagePlanId),
			data.Set("service_id", serviceId),
			data.Set("environment", environment),
			data.Set("bind_type", bindType),
			data.Set("api_id", apiId),
		} {
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, plan := range plans {
		if *plan.UsagePlanId == usagePlanId && *plan.Environment == environment {
			if bindType == API_GATEWAY_TYPE_API {
				if plan.ApiId != nil && *plan.ApiId == apiId {
					return setData()
				}
			} else {
				return setData()
			}
		}
	}
	data.SetId("")
	return nil
}

func resourceTencentCloudAPIGatewayUsagePlanAttachmentDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_usage_plan_attachment.delete")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		idMap = make(map[string]string)

		outErr, inErr error
	)

	if outErr = json.Unmarshal([]byte(data.Id()), &idMap); outErr != nil {
		return fmt.Errorf("id is broken,%s", outErr.Error())
	}

	var (
		usagePlanId = idMap["usage_plan_id"]
		serviceId   = idMap["service_id"]
		environment = idMap["environment"]
		bindType    = idMap["bind_type"]
		apiId       = idMap["api_id"]
	)
	if usagePlanId == "" || serviceId == "" || environment == "" || bindType == "" {
		return fmt.Errorf("id is broken")
	}
	if bindType == API_GATEWAY_TYPE_API && apiId == "" {
		return fmt.Errorf("id is broken")
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr = apiGatewayService.UnBindEnvironment(ctx,
			serviceId,
			usagePlanId,
			environment,
			bindType,
			apiId)

		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	return nil
}
