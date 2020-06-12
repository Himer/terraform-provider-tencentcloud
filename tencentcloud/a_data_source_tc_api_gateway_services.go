/*
Use this data source to query api gateway services.

Example Usage

```hcl
resource "tencentcloud_api_gateway_service" "service" {
  service_name = "niceservice"
  protocol     = "http&https"
  service_desc = "your nice service"
  net_type     = ["INNER", "OUTER"]
  ip_version   = "IPv4"
}

data "tencentcloud_api_gateway_services" "name" {
    service_name = tencentcloud_api_gateway_service.service.service_name
}

data "tencentcloud_api_gateway_services" "ids" {
    service_id = tencentcloud_api_gateway_service.service.id
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

func dataSourceTencentCloudAPIGatewayServices() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudAPIGatewayServicesRead,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service name for query.",
			},
			"service_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service id for query.",
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
				Description: "A list of services. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Custom service id.",
						},
						"service_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Custom service name.",
						},
						"protocol": {
							Type:         schema.TypeString,
							Computed:    true,
							Description:  "Service frontend request type, such as `http`, `https`, and `http&https`.",
						},
						"service_desc": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Custom service description.",
						},
						"exclusive_set_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Self-deployed cluster name, which is used to specify the self-deployed cluster where the service is to be created.",
						},
						"net_type": {
							Type:     schema.TypeSet,
							Computed:    true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Network type list, which is used to specify the supported network types. " +
								"`INNER` indicates access over private network, and `OUTER` indicates access over public network.",
						},
						"ip_version": {
							Type:         schema.TypeString,
							Computed:    true,
							Description:  "IP version number. Valid values: `IPv4` (default value), `IPv6`.",
						},
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
						"usage_plan_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of attach usage plans. Each element contains the following attributes:",
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
									"bind_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Binding type.",
									},
									"api_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of the api.",
									},
								},
							},
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
				},
			},
		},
	}
}

func dataSourceTencentCloudAPIGatewayServicesRead(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_api_gateway_services.read")()

	var (
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		serviceName  = data.Get("service_name").(string)
		serviceId = data.Get("service_id").(string)
		services   []*apigateway.Service

		has bool
		inErr ,outErr error
	)

	if outErr := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		services, inErr = apiGatewayService.DescribeServicesStatus(ctx,serviceId,serviceName)
		if inErr != nil {
			return retryError(inErr,InternalError)
		}
		return nil
	}); outErr != nil {
		return outErr
	}

	list := make([]map[string]interface{}, 0, len(services))

	for _, service := range services {

		var info apigateway.DescribeServiceResponse

		if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
			info, has, inErr = apiGatewayService.DescribeService(ctx, *service.ServiceId)
			if inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
		if !has{
			continue
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

		var plans []*apigateway.ApiUsagePlan

		var planList = make([]map[string]interface{}, 0, len(info.Response.ApiIdStatusSet))
		var hasContains = make(map[string]bool)

		//from service
		if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
			plans, inErr = apiGatewayService.DescribeServiceUsagePlan(ctx, *service.ServiceId)
			if inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}

		for _, item := range plans {
			if hasContains[*item.UsagePlanId] {
				continue
			}
			hasContains[*item.UsagePlanId] = true
			planList = append(
				planList, map[string]interface{}{
					"usage_plan_id":   item.UsagePlanId,
					"usage_plan_name": item.UsagePlanName,
					"bind_type":       API_GATEWAY_TYPE_SERVICE,
					"api_id":"",
				})
		}

		//from api
		if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
			plans, inErr = apiGatewayService.DescribeApiUsagePlan(ctx, *service.ServiceId)
			if inErr != nil {
				return retryError(inErr, InternalError)
			}
			return nil
		}); outErr != nil {
			return outErr
		}
		for _, item := range plans {
			planList = append(
				planList, map[string]interface{}{
					"usage_plan_id":   item.UsagePlanId,
					"usage_plan_name": item.UsagePlanName,
					"bind_type":       API_GATEWAY_TYPE_API,
					"api_id": item.ApiId,
				})
		}

		list = append(list, map[string]interface{}{
			"service_id":info.Response.ServiceId,
			"service_name":info.Response.ServiceName,
			"protocol":info.Response.Protocol,
			"service_desc":info.Response.ServiceDesc,
			"exclusive_set_name":info.Response.ExclusiveSetName,
			"ip_version":info.Response.IpVersion,
			"net_type": info.Response.NetTypes,
			"internal_sub_domain":info.Response.InternalSubDomain,
			"outer_sub_domain":info.Response.OuterSubDomain,
			"inner_http_port":info.Response.InnerHttpPort,
			"inner_https_port":info.Response.InnerHttpsPort,
			"modify_time":info.Response.ModifiedTime,
			"create_time":info.Response.CreatedTime,
			"api_list":apiList,
			"usage_plan_list":planList,
		})
	}

	byteId, err := json.Marshal(map[string]interface{}{
		"service_name":   serviceName,
		"service_id": serviceId,
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
