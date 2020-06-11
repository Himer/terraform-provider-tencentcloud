/*
Use this resource to create api gateway access key.

Example Usage

```hcl
resource "tencentcloud_api_gateway_api_key" "test" {
  secret_name = "my_api_key"
  status      = "on"
}
```

Import

api gateway access key can be imported using the id, e.g.

```
$ terraform import tencentcloud_api_gateway_api_key.test AKIDMZwceezso9ps5p8jkro8a9fwe1e7nzF2k50B
```

*/
package tencentcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
)

func resourceTencentCloudAPIGatewayAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayAPIKeyCreate,
		Read:   resourceTencentCloudAPIGatewayAPIKeyRead,
		Update: resourceTencentCloudAPIGatewayAPIKeyUpdate,
		Delete: resourceTencentCloudAPIGatewayAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Custom key name.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      API_GATEWAY_KEY_ENABLED,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_KEYS),
				Description:  "Key status. `on` or `off`.",
			},

			// Computed values.
			"access_key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Created api key id, This field is exactly the same as id.",
			},
			"access_key_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Created api key.",
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
	}
}

func resourceTencentCloudAPIGatewayAPIKeyCreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api_key.create")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		secretName = data.Get("secret_name").(string)
		statusStr  = data.Get("status").(string)

		accessKeyId   string
		inErr, outErr error
	)

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		accessKeyId, inErr = apiGatewayService.CreateApiKey(ctx, secretName)
		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}

	data.SetId(accessKeyId)

	//wait api key create ok
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr := apiGatewayService.DescribeApiKey(ctx, accessKeyId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		if has {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("accessKeyId %s not found on server", accessKeyId))

	}); outErr != nil {
		return outErr
	}

	//set status to disable
	if statusStr == API_GATEWAY_KEY_DISABLED {
		outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			if inErr = apiGatewayService.DisableApiKey(ctx, accessKeyId); inErr != nil {
				return retryError(inErr,InternalError)
			}
			return nil
		})
	}
	return resourceTencentCloudAPIGatewayAPIKeyRead(data, meta)

}
func resourceTencentCloudAPIGatewayAPIKeyRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api_key.create")()
	defer inconsistentCheck(data, meta)()
	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		accessKeyId = data.Id()

		inErr  error
		apiKey *apigateway.ApiKey
		has    bool
	)

	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		apiKey, has, inErr = apiGatewayService.DescribeApiKey(ctx, accessKeyId)
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

	errs := []error{
		data.Set("secret_name", apiKey.SecretName),
		data.Set("status", API_GATEWAY_KEY_INT2STRS[*apiKey.Status]),
		data.Set("access_key_id", apiKey.AccessKeyId),
		data.Set("access_key_secret", apiKey.AccessKeySecret),
		data.Set("modify_time", apiKey.ModifiedTime),
		data.Set("create_time", apiKey.CreatedTime),
	}

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
func resourceTencentCloudAPIGatewayAPIKeyUpdate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api_key.update")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		accessKeyId       = data.Id()
	)

	if data.HasChange("status") {
		var (
			statusStr = data.Get("status").(string)
			inErr     error
		)

		if outErr := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			if statusStr == API_GATEWAY_KEY_DISABLED {
				inErr = apiGatewayService.DisableApiKey(ctx, accessKeyId)
			} else {
				inErr = apiGatewayService.EnableApiKey(ctx, accessKeyId)
			}
			if inErr != nil {
				return retryError(inErr,InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
	}

	return resourceTencentCloudAPIGatewayAPIKeyRead(data, meta)
}
func resourceTencentCloudAPIGatewayAPIKeyDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_api_key.delete")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		accessKeyId       = data.Id()
	)

	//set status to disable before delete
	if data.Get("status") != API_GATEWAY_KEY_DISABLED {
		if outErr := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			if inErr := apiGatewayService.DisableApiKey(ctx, accessKeyId); inErr != nil {
				return retryError(inErr,InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
	}

	return resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr := apiGatewayService.DeleteApiKey(ctx, accessKeyId)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	})
}
