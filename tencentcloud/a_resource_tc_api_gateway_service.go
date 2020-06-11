/*
Use this resource to create api gateway service.

Example Usage

```hcl
resource "tencentcloud_api_gateway_service" "service" {
  service_name = "niceservice"
  protocol     = "http&https"
  service_desc = "your nice service"
  net_type     = ["INNER", "OUTER"]
  ip_version   = "IPv4"
}
```

Import

api gateway service can be imported using the id, e.g.

```
$ terraform import tencentcloud_api_gateway_service.service service-pg6ud8pa
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

func resourceTencentCloudAPIGatewayService() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayServiceCreate,
		Read:   resourceTencentCloudAPIGatewayServiceRead,
		Update: resourceTencentCloudAPIGatewayServiceUpdate,
		Delete: resourceTencentCloudAPIGatewayServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Custom service name.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_SERVICE_PROTOCOLS),
				Description:  "Service frontend request type, such as `http`, `https`, and `http&https`.",
			},
			"service_desc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom service description.",
			},
			"exclusive_set_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Self-deployed cluster name, which is used to specify the self-deployed cluster where the service is to be created.",
			},
			"net_type": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: "Network type list, which is used to specify the supported network types. " +
					"`INNER` indicates access over private network, and `OUTER` indicates access over public network.",
			},
			"ip_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_NET_IP_VERSIONS),
				Description:  "IP version number. Valid values: `IPv4` (default value), `IPv6`.",
			},
			"set_server_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Cluster name, which is reserved and used by the tsf serverless type.",
			},
			"appid_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "User type, which is reserved and can be used by serverless users.",
			},

			// Computed values.
			"internal_sub_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Private network access subdomain name.",
			},
			"outer_sub_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Public network access subdomain name.",
			},
			"inner_http_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number for http access over private network.",
			},
			"inner_https_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number for https access over private network.",
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
			"api_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of apis. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the api.",
						},
						"api_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the api",
						},
						"api_desc": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the api.",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Path of the api.",
						},
						"method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Method of the api.",
						},
					},
				},
			},
		},
	}
}

func resourceTencentCloudAPIGatewayServiceCreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_service.create")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		serviceName      = data.Get("service_name").(string)
		protocol         = data.Get("protocol").(string)
		serviceDesc      = data.Get("service_desc").(string)
		exclusiveSetName = data.Get("exclusive_set_name").(string)
		ipVersion        = data.Get("ip_version").(string)
		setServerName    = data.Get("set_server_name").(string)
		appidType        = data.Get("appid_type").(string)
		netTypes         = helper.InterfacesStrings(data.Get("net_type").(*schema.Set).List())

		serviceId     string
		inErr, outErr error
	)

	for _, v := range netTypes {
		if v != API_GATEWAY_NET_TYPE_INNER && v != API_GATEWAY_NET_TYPE_OUTER {
			return fmt.Errorf("not support appid_type %s", v)
		}
	}
	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		serviceId, inErr = apiGatewayService.CreateService(ctx,
			serviceName,
			protocol,
			serviceDesc,
			exclusiveSetName,
			ipVersion,
			setServerName,
			appidType,
			netTypes)

		if inErr != nil {
			return retryError(inErr)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	data.SetId(serviceId)

	//wait service create ok
	if outErr := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		_, has, inErr := apiGatewayService.DescribeService(ctx, serviceId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		if has {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("service %s not found on server", serviceId))

	}); outErr != nil {
		return outErr
	}
	return resourceTencentCloudAPIGatewayServiceRead(data, meta)

}
func resourceTencentCloudAPIGatewayServiceRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_service.create")()
	defer inconsistentCheck(data, meta)()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		info apigateway.DescribeServiceResponse
		has  bool

		serviceId     = data.Id()
		inErr, outErr error
	)

	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeService(ctx, serviceId)
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

	var apiList = make([]map[string]interface{}, 0, len(info.Response.ApiIdStatusSet))

	for _, item := range info.Response.ApiIdStatusSet {
		apiList = append(
			apiList, map[string]interface{}{
				"api_id":   item.ApiId,
				"api_name": item.ApiName,
				"api_desc": item.ApiDesc,
				"path":     item.Path,
				"method":   item.Method,
			})
	}
	errs := []error{
		data.Set("service_name", info.Response.ServiceName),
		data.Set("protocol", info.Response.Protocol),
		data.Set("service_desc", info.Response.ServiceDesc),
		data.Set("exclusive_set_name", info.Response.ExclusiveSetName),
		data.Set("ip_version", info.Response.IpVersion),
		data.Set("net_type", info.Response.NetTypes),
		data.Set("internal_sub_domain", info.Response.InternalSubDomain),
		data.Set("outer_sub_domain", info.Response.OuterSubDomain),
		data.Set("inner_http_port", info.Response.InnerHttpPort),
		data.Set("inner_https_port", info.Response.InnerHttpsPort),
		data.Set("modify_time", info.Response.ModifiedTime),
		data.Set("create_time", info.Response.CreatedTime),
		data.Set("api_list", apiList),
	}
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
func resourceTencentCloudAPIGatewayServiceUpdate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_service.update")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		serviceName = data.Get("service_name").(string)
		protocol    = data.Get("protocol").(string)
		serviceDesc = data.Get("service_desc").(string)
		netTypes    = helper.InterfacesStrings(data.Get("net_type").(*schema.Set).List())

		serviceId     = data.Id()
		inErr, outErr error
	)

	for _, v := range netTypes {
		if v != API_GATEWAY_NET_TYPE_INNER && v != API_GATEWAY_NET_TYPE_OUTER {
			return fmt.Errorf("not support appid_type %s", v)
		}
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		if inErr = apiGatewayService.ModifyService(ctx,
			serviceId,
			serviceName,
			protocol,
			serviceDesc,
			netTypes); inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	return resourceTencentCloudAPIGatewayServiceRead(data, meta)
}

func resourceTencentCloudAPIGatewayServiceDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_service.delete")()
	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		serviceId         = data.Id()
		inErr, outErr     error
	)

	for _,env:=range API_GATEWAY_SERVICE_ENVS{
		outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			if inErr = apiGatewayService.UnReleaseService(ctx,
				serviceId,
				env); inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		})
		if outErr != nil {
			return outErr
		}
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		if inErr = apiGatewayService.DeleteService(ctx,
			serviceId); inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}
	return nil
}
