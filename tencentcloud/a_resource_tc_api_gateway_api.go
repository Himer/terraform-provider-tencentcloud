/*
Use this resource to create api gateway api.

Example Usage

```hcl
resource "tencentcloud_api_gateway_service" "service" {
  service_name = "ck"
  protocol     = "http&https"
  service_desc = "your nice service"
  net_type     = ["INNER", "OUTER"]
  ip_version   = "IPv4"
}

resource "tencentcloud_api_gateway_api" "api" {
  service_id            = tencentcloud_api_gateway_service.service.id
  api_name              = "hello"
  api_desc              = "my hello api"
  auth_type             = "NONE"
  protocol              = "HTTP"
  enable_cors           = true
  request_config_path   = "/user/info"
  request_config_method = "GET"

  request_parameters {
    name          = "name"
    position      = "QUERY"
    type          = "string"
    desc          = "who are you?"
    default_value = "tom"
    required      = true
  }
  service_config_type      = "HTTP"
  service_config_timeout   = 15
  service_config_url       = "http://www.qq.com"
  service_config_path      = "/user"
  service_config_method    = "GET"
  response_type            = "HTML"
  response_success_example = "success"
  response_fail_example    = "fail"
  response_error_codes {
    code           = 100
    msg            = "system error"
    desc           = "system error code"
    converted_code = -100
    need_convert   = true
  }

}

resource "tencentcloud_api_gateway_api" "scf" {
  service_id                            = tencentcloud_api_gateway_service.service.id
  api_name                              = "scf_hello"
  api_desc                              = "my scf hello api"
  auth_type                             = "NONE"
  protocol                              = "HTTP"
  request_config_path                   = "/user/info2"
  request_config_method                 = "GET"
  service_config_type                   = "SCF"
  service_config_scf_function_name      = "resource-bot"
  service_config_scf_function_namespace = "default"
  service_config_scf_function_qualifier = "$LATEST"
}


resource "tencentcloud_api_gateway_api" "mock" {
  service_id                         = tencentcloud_api_gateway_service.service.id
  api_name                           = "mock_hello"
  api_desc                           = "my mock hello api"
  auth_type                          = "NONE"
  protocol                           = "HTTP"
  request_config_path                = "/user/mock"
  request_config_method              = "POST"
  service_config_type                = "MOCK"
  service_config_mock_return_message = "guaguajiao"
}

resource "tencentcloud_api_gateway_api" "websock" {
  service_id            = tencentcloud_api_gateway_service.service.id
  api_name              = "websock_hello"
  api_desc              = "my websock hello api"
  auth_type             = "NONE"
  protocol              = "WEBSOCKET"
  request_config_path   = "/user/websock"
  request_config_method = "GET"

  service_config_type    = "WEBSOCKET"
  service_config_timeout = 15
  service_config_url     = "ws://www.qq.com"
  service_config_path    = "/user"
  service_config_method  = "GET"
}
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
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/ratelimit"
	"strings"
)

func resourceTencentCloudAPIGatewayAPI() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAPIGatewayAPICreate,
		Read:   resourceTencentCloudAPIGatewayAPIRead,
		Update: resourceTencentCloudAPIGatewayAPIUpdate,
		Delete: resourceTencentCloudAPIGatewayAPIDelete,

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Which service this api belongs.Refer to resource `tencentcloud_api_gateway_service`",
			},
			"api_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Custom api name.",
			},
			"api_desc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom api description.",
			},
			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      API_GATEWAY_AUTH_TYPE_NONE,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_AUTH_TYPES),
				Description:  "API authentication type. Valid values: `SECRET` (key pair authentication),`NONE` (no authentication). Default value: `NONE`.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      API_GATEWAY_API_PROTOCOL_HTTP,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_API_PROTOCOLS),
				Description:  "API frontend request type, such as `HTTP`,`WEBSOCKET`. Default value: `HTTP`.",
			},
			"enable_cors": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to enable CORS. Default value: `true`",
			},
			"request_config_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Request frontend path configuration. Like `/user/getinfo`",
			},
			"request_config_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "GET",
				ValidateFunc: validateAllowedStringValue([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "ANY"}),
				Description:  "Request frontend method configuration. Like `GET`,`POST`,`PUT`,`DELETE`,`HEAD`,`ANY`. Default value: `GET`",
			},
			"request_parameters": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Frontend request parameters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Parameter name.",
						},
						"position": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Parameter location.",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Parameter type.",
						},
						"desc": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Parameter description.",
						},
						"default_value": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Parameter default value.",
						},
						"required": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If this parameter required. Default value: `false`",
						},
					},
				},
			},
			"service_config_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      API_GATEWAY_SERVICE_TYPE_HTTP,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_SERVICE_TYPES),
				Description:  "API backend service type. Valid values: " + strings.Join(API_GATEWAY_SERVICE_TYPES, ",") + ". Default value: `HTTP`",
			},
			"service_config_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "API backend service timeout period in seconds. Default is `5`",
			},
			"service_config_product": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Backend type. This parameter takes effect when vpc is enabled. Currently, only `clb` is supported.",
			},
			"service_config_vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Unique vpc id.",
			},
			"service_config_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "API backend service url. This parameter is required when `service_config_type` is `HTTP`.",
			},
			"service_config_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "API backend service path, such as /path. If `service_config_type` is `HTTP`, this parameter will be required. The frontend `request_config_path` and backend path `service_config_path` can be different.",
			},
			"service_config_method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "API backend service request method, such as `GET`. If `service_config_type` is `HTTP`, this parameter will be required. The frontend `request_config_method` and backend method `service_config_method` can be different.",
			},
			"service_config_scf_function_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SCF function name. This parameter takes effect when `service_config_type` is `SCF`.",
			},
			"service_config_scf_function_namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SCF function namespace. This parameter takes effect when  `service_config_type` is `SCF`.",
			},
			"service_config_scf_function_qualifier": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SCF function version. This parameter takes effect when `service_config_type`  is `SCF`.",
			},
			"service_config_mock_return_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Returned information of API backend mocking. This parameter is required when `service_config_type`  is `MOCK`.",
			},
			"response_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAllowedStringValue(API_GATEWAY_API_RESPONSE_TYPES),
				Description:  "Return type. Valid values: " + strings.Join(API_GATEWAY_API_RESPONSE_TYPES, ",") + ". Default value: `HTML`.",
			},
			"response_success_example": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Successful response sample of custom response configuration.",
			},
			"response_fail_example": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Response failure sample of custom response configuration.",
			},
			"response_error_codes": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Custom error code configuration. Must keep at least one after set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Custom response configuration error code.",
						},
						"msg": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Custom response configuration error message.",
						},
						"desc": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Parameter description.",
						},
						"converted_code": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Custom error code conversion.",
						},
						"need_convert": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to enable error code conversion. Default value: `false`",
						},
					},
				},
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
		},
	}
}

func resourceTencentCloudAPIGatewayAPICreate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api.create")()

	var (
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)
		outErr, inErr     error
		response          = apigateway.NewCreateApiResponse()
		request           = apigateway.NewCreateApiRequest()
		serviceId         = data.Get("service_id").(string)
		has               bool
	)

	request.ServiceId = &serviceId
	request.ApiName = helper.String(data.Get("api_name").(string))
	if object, ok := data.GetOk("api_desc"); ok {
		request.ApiDesc = helper.String(object.(string))
	}

	request.AuthType = helper.String(data.Get("auth_type").(string))
	request.Protocol = helper.String(data.Get("protocol").(string))
	request.EnableCORS = helper.Bool(data.Get("enable_cors").(bool))
	request.RequestConfig =
		&apigateway.ApiRequestConfig{Path: helper.String(data.Get("request_config_path").(string)),
			Method: helper.String(data.Get("request_config_method").(string))}

	if object, ok := data.GetOk("request_parameters"); ok {
		parameters := object.(*schema.Set).List()
		request.RequestParameters = make([]*apigateway.RequestParameter, 0, len(parameters))
		for _, parameter := range parameters {
			parameterMap := parameter.(map[string]interface{})
			requestParameter := &apigateway.RequestParameter{}
			requestParameter.Name = helper.String(parameterMap["name"].(string))
			requestParameter.Position = helper.String(parameterMap["position"].(string))
			requestParameter.Type = helper.String(parameterMap["type"].(string))
			requestParameter.Required = helper.Bool(parameterMap["required"].(bool))
			if parameterMap["desc"] != nil {
				requestParameter.Desc = helper.String(parameterMap["desc"].(string))
			}
			if parameterMap["default_value"] != nil {
				requestParameter.DefaultValue = helper.String(parameterMap["default_value"].(string))
			}
			request.RequestParameters = append(request.RequestParameters, requestParameter)
		}
	}

	var serviceType = data.Get("service_config_type").(string)
	request.ServiceType = &serviceType
	request.ServiceTimeout = helper.IntInt64(data.Get("service_config_timeout").(int))

	switch serviceType {

	case API_GATEWAY_SERVICE_TYPE_WEBSOCKET, API_GATEWAY_SERVICE_TYPE_HTTP:
		serviceConfigProduct := data.Get("service_config_product").(string)
		serviceConfigVpcId := data.Get("service_config_vpc_id").(string)
		serviceConfigUrl := data.Get("service_config_url").(string)
		serviceConfigPath := data.Get("service_config_path").(string)
		serviceConfigMethod := data.Get("service_config_method").(string)
		if serviceConfigProduct != "" {
			if serviceConfigProduct != "clb" {
				return fmt.Errorf("`service_config_product` only support `clb` now")
			}
			if serviceConfigVpcId == "" {
				return fmt.Errorf("`service_config_product` need param `service_config_vpc_id`")
			}
		}
		if serviceConfigUrl == "" || serviceConfigPath == "" || serviceConfigMethod == "" {
			return fmt.Errorf("`service_config_url`,`service_config_path`,`service_config_method` is needed if `service_config_type` is `WEBSOCKET` or `HTTP`")
		}
		request.ServiceConfig = &apigateway.ServiceConfig{}
		if serviceConfigProduct != "" {
			request.ServiceConfig.Product = &serviceConfigProduct
		}
		if serviceConfigVpcId != "" {
			request.ServiceConfig.UniqVpcId = &serviceConfigVpcId
		}

		request.ServiceConfig.Url = &serviceConfigUrl
		request.ServiceConfig.Path = &serviceConfigPath
		request.ServiceConfig.Method = &serviceConfigMethod

	case API_GATEWAY_SERVICE_TYPE_MOCK:
		serviceConfigMockReturnMessage := data.Get("service_config_mock_return_message").(string)
		if serviceConfigMockReturnMessage == "" {
			return fmt.Errorf("`service_config_mock_return_message` is needed if `service_config_type` is `MOCK`")
		}
		request.ServiceMockReturnMessage = &serviceConfigMockReturnMessage

	case API_GATEWAY_SERVICE_TYPE_SCF:
		scfFunctionName := data.Get("service_config_scf_function_name").(string)
		scfFunctionNamespace := data.Get("service_config_scf_function_namespace").(string)
		scfFunctionQualifier := data.Get("service_config_scf_function_qualifier").(string)
		if scfFunctionName == "" || scfFunctionNamespace == "" || scfFunctionQualifier == "" {
			return fmt.Errorf("`service_config_scf_function_name`,`service_config_scf_function_namespace`,`service_config_scf_function_qualifier` is needed if `service_config_type` is `SCF`")
		}
		request.ServiceScfFunctionName = &scfFunctionName
		request.ServiceScfFunctionNamespace = &scfFunctionNamespace
		request.ServiceScfFunctionQualifier = &scfFunctionQualifier
	}

	request.ResponseType = helper.String(data.Get("response_type").(string))

	if object, ok := data.GetOk("response_success_example"); ok {
		request.ResponseSuccessExample = helper.String(object.(string))
	}

	if object, ok := data.GetOk("response_fail_example"); ok {
		request.ResponseFailExample = helper.String(object.(string))
	}
	if object, ok := data.GetOk("response_error_codes"); ok {
		codes := object.(*schema.Set).List()
		request.ResponseErrorCodes = make([]*apigateway.ResponseErrorCodeReq, 0, len(codes))
		for _, code := range codes {
			codeMap := code.(map[string]interface{})
			codeReq := &apigateway.ResponseErrorCodeReq{}
			codeReq.Code = helper.IntInt64(codeMap["code"].(int))
			codeReq.Msg = helper.String(codeMap["msg"].(string))

			if codeMap["desc"] != nil {
				codeReq.Desc = helper.String(codeMap["desc"].(string))
			}
			if codeMap["converted_code"] != nil {
				codeReq.ConvertedCode = helper.IntInt64(codeMap["converted_code"].(int))
			}
			if codeMap["need_convert"] != nil {
				codeReq.NeedConvert = helper.Bool(codeMap["need_convert"].(bool))
			}
			if *codeReq.NeedConvert && codeReq.ConvertedCode == nil {
				return fmt.Errorf("`need_convert` need `converted_code`setted")
			}
			request.ResponseErrorCodes = append(request.ResponseErrorCodes, codeReq)
		}
	}

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
		return fmt.Errorf("service %s not exist on server", serviceId)
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		response, inErr = apiGatewayService.client.UseAPIGatewayClient().CreateApi(request)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}

	if response == nil || response.Response.Result == nil || response.Response.Result.ApiId == nil {
		return fmt.Errorf("create api fail,return nil response")
	}

	data.SetId(*response.Response.Result.ApiId)

	return resourceTencentCloudAPIGatewayAPIRead(data, meta)

}
func resourceTencentCloudAPIGatewayAPIRead(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api.read")()
	defer inconsistentCheck(data, meta)()

	var (
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)

		outErr, inErr error

		apiId     = data.Id()
		serviceId = data.Get("service_id").(string)

		info apigateway.ApiInfo
		has  bool
	)

	if outErr = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		info, has, inErr = apiGatewayService.DescribeApi(ctx, serviceId, apiId)
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

	errs := []error{
		data.Set("service_id", info.ServiceId),
		data.Set("api_name", info.ApiName),
		data.Set("api_desc", info.ApiDesc),
		data.Set("auth_type", info.AuthType),
		data.Set("protocol", info.Protocol),
		data.Set("enable_cors", info.EnableCORS),
		data.Set("response_type", info.ResponseType),
		data.Set("response_success_example", info.ResponseSuccessExample),
		data.Set("response_fail_example", info.ResponseFailExample),
		data.Set("service_config_type", info.ServiceType),
		data.Set("service_config_timeout", info.ServiceTimeout),
		data.Set("service_config_scf_function_name", info.ServiceScfFunctionName),
		data.Set("service_config_scf_function_namespace", info.ServiceScfFunctionNamespace),
		data.Set("service_config_scf_function_qualifier", info.ServiceScfFunctionQualifier),
		data.Set("service_config_mock_return_message", info.ServiceMockReturnMessage),
		data.Set("modify_time", info.ModifiedTime),
		data.Set("create_time", info.CreatedTime),
	}

	if info.RequestConfig != nil {
		errs = append(errs,
			data.Set("request_config_path", info.RequestConfig.Path),
			data.Set("request_config_method", info.RequestConfig.Method))
	}
	if info.RequestParameters != nil {
		list := make([]map[string]interface{}, 0, len(info.RequestParameters))
		for _, param := range info.RequestParameters {
			list = append(list, map[string]interface{}{
				"name":          param.Name,
				"position":      param.Position,
				"type":          param.Type,
				"desc":          param.Desc,
				"default_value": param.DefaultValue,
				"required":      param.Required,
			})
		}
		errs = append(errs, data.Set("request_parameters", list))
	}

	if info.ServiceConfig != nil {
		errs = append(errs,
			data.Set("service_config_product", info.ServiceConfig.Product),
			data.Set("service_config_vpc_id", info.ServiceConfig.UniqVpcId),
			data.Set("service_config_url", info.ServiceConfig.Url),
			data.Set("service_config_path", info.ServiceConfig.Path),
			data.Set("service_config_method", info.ServiceConfig.Method))
	}

	if info.ResponseErrorCodes != nil {
		list := make([]map[string]interface{}, 0, len(info.ResponseErrorCodes))
		for _, code := range info.ResponseErrorCodes {
			list = append(list, map[string]interface{}{
				"code":           code.Code,
				"msg":            code.Msg,
				"desc":           code.Desc,
				"converted_code": code.ConvertedCode,
				"need_convert":   code.NeedConvert,
			})
		}
		errs = append(errs, data.Set("response_error_codes", list))
	}

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
func resourceTencentCloudAPIGatewayAPIUpdate(data *schema.ResourceData, meta interface{}) error {

	defer logElapsed("resource.tencentcloud_api_gateway_api.update")()

	var (
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}

		outErr, inErr error
		response      = apigateway.NewModifyApiResponse()
		request       = apigateway.NewModifyApiRequest()

		apiId     = data.Id()
		serviceId = data.Get("service_id").(string)
	)

	request.ServiceId = &serviceId
	request.ApiId = &apiId
	request.ApiName = helper.String(data.Get("api_name").(string))
	if object, ok := data.GetOk("api_desc"); ok {
		request.ApiDesc = helper.String(object.(string))
	}

	request.AuthType = helper.String(data.Get("auth_type").(string))
	request.Protocol = helper.String(data.Get("protocol").(string))
	request.EnableCORS = helper.Bool(data.Get("enable_cors").(bool))
	request.RequestConfig =
		&apigateway.RequestConfig{Path: helper.String(data.Get("request_config_path").(string)),
			Method: helper.String(data.Get("request_config_method").(string))}

	if object, ok := data.GetOk("request_parameters"); ok {
		parameters := object.(*schema.Set).List()
		request.RequestParameters = make([]*apigateway.ReqParameter, 0, len(parameters))
		for _, parameter := range parameters {
			parameterMap := parameter.(map[string]interface{})
			requestParameter := &apigateway.ReqParameter{}
			requestParameter.Name = helper.String(parameterMap["name"].(string))
			requestParameter.Position = helper.String(parameterMap["position"].(string))
			requestParameter.Type = helper.String(parameterMap["type"].(string))
			requestParameter.Required = helper.Bool(parameterMap["required"].(bool))
			if parameterMap["desc"] != nil {
				requestParameter.Desc = helper.String(parameterMap["desc"].(string))
			}
			if parameterMap["default_value"] != nil {
				requestParameter.DefaultValue = helper.String(parameterMap["default_value"].(string))
			}
			request.RequestParameters = append(request.RequestParameters, requestParameter)
		}
	}

	var serviceType = data.Get("service_config_type").(string)
	request.ServiceType = &serviceType
	request.ServiceTimeout = helper.IntInt64(data.Get("service_config_timeout").(int))

	switch serviceType {

	case API_GATEWAY_SERVICE_TYPE_WEBSOCKET, API_GATEWAY_SERVICE_TYPE_HTTP:
		serviceConfigProduct := data.Get("service_config_product").(string)
		serviceConfigVpcId := data.Get("service_config_vpc_id").(string)
		serviceConfigUrl := data.Get("service_config_url").(string)
		serviceConfigPath := data.Get("service_config_path").(string)
		serviceConfigMethod := data.Get("service_config_method").(string)
		if serviceConfigProduct != "" {
			if serviceConfigProduct != "clb" {
				return fmt.Errorf("`service_config_product` only support `clb` now")
			}
			if serviceConfigVpcId == "" {
				return fmt.Errorf("`service_config_product` need param `service_config_vpc_id`")
			}
		}
		if serviceConfigUrl == "" || serviceConfigPath == "" || serviceConfigMethod == "" {
			return fmt.Errorf("`service_config_url`,`service_config_path`,`service_config_method` is needed if `service_config_type` is `WEBSOCKET` or `HTTP`")
		}
		request.ServiceConfig = &apigateway.ServiceConfig{}
		if serviceConfigProduct != "" {
			request.ServiceConfig.Product = &serviceConfigProduct
		}
		if serviceConfigVpcId != "" {
			request.ServiceConfig.UniqVpcId = &serviceConfigVpcId
		}
		request.ServiceConfig.Url = &serviceConfigUrl
		request.ServiceConfig.Path = &serviceConfigPath
		request.ServiceConfig.Method = &serviceConfigMethod

	case API_GATEWAY_SERVICE_TYPE_MOCK:
		serviceConfigMockReturnMessage := data.Get("service_config_mock_return_message").(string)
		if serviceConfigMockReturnMessage == "" {
			return fmt.Errorf("`service_config_mock_return_message` is needed if `service_config_type` is `MOCK`")
		}
		request.ServiceMockReturnMessage = &serviceConfigMockReturnMessage

	case API_GATEWAY_SERVICE_TYPE_SCF:
		scfFunctionName := data.Get("service_config_scf_function_name").(string)
		scfFunctionNamespace := data.Get("service_config_scf_function_namespace").(string)
		scfFunctionQualifier := data.Get("service_config_scf_function_qualifier").(string)
		if scfFunctionName == "" || scfFunctionNamespace == "" || scfFunctionQualifier == "" {
			return fmt.Errorf("`service_config_scf_function_name`,`service_config_scf_function_namespace`,`service_config_scf_function_qualifier` is needed if `service_config_type` is `SCF`")
		}
		request.ServiceScfFunctionName = &scfFunctionName
		request.ServiceScfFunctionNamespace = &scfFunctionNamespace
		request.ServiceScfFunctionQualifier = &scfFunctionQualifier
	}

	request.ResponseType = helper.String(data.Get("response_type").(string))

	if object, ok := data.GetOk("response_success_example"); ok {
		request.ResponseSuccessExample = helper.String(object.(string))
	}

	if object, ok := data.GetOk("response_fail_example"); ok {
		request.ResponseFailExample = helper.String(object.(string))
	}

	oldInterface, newInterface := data.GetChange("response_error_codes")

	if oldInterface.(*schema.Set).Len() > 0 && newInterface.(*schema.Set).Len() == 0 {
		return fmt.Errorf("`response_error_codes` must keep at least one after set")
	}

	if object, ok := data.GetOk("response_error_codes"); ok {
		codes := object.(*schema.Set).List()
		request.ResponseErrorCodes = make([]*apigateway.ResponseErrorCodeReq, 0, len(codes))
		for _, code := range codes {
			codeMap := code.(map[string]interface{})
			codeReq := &apigateway.ResponseErrorCodeReq{}
			codeReq.Code = helper.IntInt64(codeMap["code"].(int))
			codeReq.Msg = helper.String(codeMap["msg"].(string))

			if codeMap["desc"] != nil {
				codeReq.Desc = helper.String(codeMap["desc"].(string))
			}
			if codeMap["converted_code"] != nil {
				codeReq.ConvertedCode = helper.IntInt64(codeMap["converted_code"].(int))
			}
			if codeMap["need_convert"] != nil {
				codeReq.NeedConvert = helper.Bool(codeMap["need_convert"].(bool))
			}
			if *codeReq.NeedConvert && codeReq.ConvertedCode == nil {
				return fmt.Errorf("`need_convert` need `converted_code`setted")
			}
			request.ResponseErrorCodes = append(request.ResponseErrorCodes, codeReq)
		}
	}

	outErr = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		response, inErr = apiGatewayService.client.UseAPIGatewayClient().ModifyApi(request)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	})
	if outErr != nil {
		return outErr
	}

	if response == nil {
		return fmt.Errorf("modify api fail,return nil response")
	}
	return resourceTencentCloudAPIGatewayAPIRead(data, meta)
}
func resourceTencentCloudAPIGatewayAPIDelete(data *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_api_gateway_api.delete")()
	var (
		apiGatewayService = APIGatewayService{client: meta.(*TencentCloudClient).apiV3Conn}
		logId             = getLogId(contextNil)
		ctx               = context.WithValue(context.TODO(), logIdKey, logId)

		apiId     = data.Id()
		serviceId = data.Get("service_id").(string)
	)

	return resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		inErr := apiGatewayService.DeleteApi(ctx, serviceId, apiId)
		if inErr != nil {
			return retryError(inErr, InternalError)
		}
		return nil
	})

}
