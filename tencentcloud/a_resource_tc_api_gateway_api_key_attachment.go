/*
Use this resource to  api gateway attach access key to usage plan.

Example Usage

```hcl
resource "tencentcloud_api_gateway_api_key" "key" {
  secret_name = "my_api_key"
  status      = "on"
}

resource "tencentcloud_api_gateway_usage_plan" "plan" {
  usage_plan_name         = "my_plan"
  usage_plan_desc         = "nice plan"
  max_request_num         = 100
  max_request_num_pre_sec = 10
}


resource "tencentcloud_api_gateway_api_key_attachment" "attach" {
  api_key_id    = tencentcloud_api_gateway_api_key.key.id
  usage_plan_id = tencentcloud_api_gateway_usage_plan.plan.id
}
```

Import

api gateway access key can be imported using the id, e.g.

```
$ terraform import tencentcloud_api_gateway_api_key_attachment.test '{"api_key_id":"AKID110b8Rmuw7t0fP1N8bi809n327023Is7xN8f","usage_plan_id":"usagePlan-gyeafpab"}]'
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

func resourceTencentCloudAPIGatewayAPIKeyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayAPIKeyAttachmentCreate,
		Read:   resourceTencentCloudAPIGatewayAPIKeyAttachmentRead,
		Delete: resourceTencentCloudAPIGatewayAPIKeyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"api_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of api key.",
			},
			"usage_plan_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the usage plan.",
			},
		},
	}
}

func resourceTencentCloudAPIGatewayAPIKeyAttachmentCreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api_key_attachment.create")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		apiKeyId          = data.Get("api_key_id").(string)
		usagePlanId       = data.Get("usage_plan_id").(string)
		has               bool
		inErr, outErr     error
	)

	byteId, err := json.Marshal(map[string]interface{}{
		"api_key_id":    apiKeyId,
		"usage_plan_id": usagePlanId,
	})
	if err != nil {
		return err
	}

	//check usage plan is exist
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	if !has {
		return fmt.Errorf("usage plan %s is not exist", usagePlanId)
	}

	//check api key is exist
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr = apiGatewayService.DescribeApiKey(ctx, apiKeyId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}
	if !has {
		return fmt.Errorf("api key %s is not exist", apiKeyId)
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		if inErr = apiGatewayService.BindSecretId(ctx, usagePlanId, apiKeyId); inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	data.SetId(string(byteId))

	//waiting bind success
	var info apigateway.UsagePlanInfo
	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		if !has {
			return nil
		}
		for _, v := range info.BindSecretIds {
			if *v == apiKeyId {
				return nil
			}
		}
		return resource.RetryableError(
			fmt.Errorf("api key  %s attach to usage plan %s still is doing",
				apiKeyId, usagePlanId))

	}); outErr != nil {
		return outErr
	}
	if !has {
		return fmt.Errorf("usage plan %s has been deleted", usagePlanId)
	}

	data.SetId(string(byteId))

	return resourceTencentCloudAPIGatewayAPIKeyAttachmentRead(data, meta)

}
func resourceTencentCloudAPIGatewayAPIKeyAttachmentRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api_key_attachment.create")()
	defer inconsistentCheck(data, meta)()
	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		info apigateway.UsagePlanInfo

		inErr, outErr error
		has           bool
	)

	var idMaps = make(map[string]string, 2)
	if outErr = json.Unmarshal([]byte(data.Id()), &idMaps); outErr != nil {
		return fmt.Errorf("id is broken,%s", outErr.Error())
	}
	apiKeyId := idMaps["api_key_id"]
	usagePlanId := idMaps["usage_plan_id"]
	if apiKeyId == "" || usagePlanId == "" {
		return fmt.Errorf("id is broken,%s", data.Id())
	}

	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	if !has {
		data.SetId("")
		return nil
	}
	for _, v := range info.BindSecretIds {
		if *v == apiKeyId {
			if outErr = data.Set("api_key_id", apiKeyId); outErr != nil {
				return outErr
			}
			return data.Set("usage_plan_id", usagePlanId)
		}
	}
	data.SetId("")
	return nil

}

func resourceTencentCloudAPIGatewayAPIKeyAttachmentDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_api_key_attachment.delete")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		info apigateway.UsagePlanInfo

		inErr, outErr error
		has           bool
	)
	var idMaps = make(map[string]string, 2)
	if outErr = json.Unmarshal([]byte(data.Id()), &idMaps); outErr != nil {
		return fmt.Errorf("id is broken,%s", outErr.Error())
	}
	apiKeyId := idMaps["api_key_id"]
	usagePlanId := idMaps["usage_plan_id"]
	if apiKeyId == "" || usagePlanId == "" {
		return fmt.Errorf("id is broken,%s", data.Id())
	}

	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		inErr = apiGatewayService.UnBindSecretId(ctx, usagePlanId, apiKeyId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	//waiting delete ok
	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeUsagePlan(ctx, usagePlanId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		if !has {
			return nil
		}
		for _, v := range info.BindSecretIds {
			if *v == apiKeyId {
				return resource.RetryableError(
					fmt.Errorf("api key  %s attach to usage plan %s still is deleteing",
						apiKeyId, usagePlanId))
			}
		}

		return nil
	}); outErr != nil {
		return outErr
	}
	return nil

}
