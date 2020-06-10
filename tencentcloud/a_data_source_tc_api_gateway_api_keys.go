/*
Use this data source to query api gateway access keys.

Example Usage

```hcl
resource "tencentcloud_api_gateway_api_key" "test_cluster" {
  secret_name = "my_api_key"
  status      = "on"
}
resource "tencentcloud_api_gateway_api_key" "test_cluster2" {
  secret_name = "my_api_key"
  status      = "on"
}
data "tencentcloud_api_gateway_api_keys" "name" {
  secret_name = tencentcloud_api_gateway_api_key.test_cluster.secret_name
}

data "tencentcloud_api_gateway_api_keys" "id" {
  access_key_id = tencentcloud_api_gateway_api_key.test_cluster.access_key_id
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

func dataSourceTencentCloudAPIGatewayAPIKeys() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudAPIGatewayAPIKeysRead,

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom key name.",
			},
			"access_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Created api key id, This field is exactly the same as id.",
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
				Description: "A list of api keys. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "API key id.",
						},
						"status": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Key status. `on` or `off`.",
						},
						"access_key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Created api key id, This field is exactly the same as `api_key_id`.",
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
				},
			},
		},
	}
}

func dataSourceTencentCloudAPIGatewayAPIKeysRead(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_api_gateway_api_keys.read")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		secretName  = data.Get("secret_name").(string)
		accessKeyId = data.Get("access_key_id").(string)
		apiKeySet   []*apigateway.ApiKey

		inErr error
	)

	if outErr := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		apiKeySet, inErr = apiGatewayService.DescribeApiKeysStatus(ctx, secretName, accessKeyId)
		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	list := make([]map[string]interface{}, 0, len(apiKeySet))
	for _, apiKey := range apiKeySet {
		list = append(list, map[string]interface{}{
			"api_key_id":        apiKey.AccessKeyId,
			"status":            API_GATEWAY_KEY_INT2STRS[*apiKey.Status],
			"access_key_id":     apiKey.AccessKeyId,
			"access_key_secret": apiKey.AccessKeySecret,
			"modify_time":       apiKey.ModifiedTime,
			"create_time":       apiKey.CreatedTime,
		})
	}

	byteId, err := json.Marshal(map[string]interface{}{
		"secret_name":   secretName,
		"access_key_id": accessKeyId,
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
