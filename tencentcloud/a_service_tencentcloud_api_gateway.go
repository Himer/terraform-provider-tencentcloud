package tencentcloud

import (
	"context"
	"fmt"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/connectivity"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

type APIGatewayService struct {
	client *connectivity.TencentCloudClient
}

func (me *APIGatewayService) CreateApiKey(ctx context.Context, secretName string) (accessKeyId string, errRet error) {
	request := apigateway.NewCreateApiKeyRequest()
	request.SecretName = &secretName
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseAPIGatewayClient().CreateApiKey(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil || response.Response.Result.AccessKeyId == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty AccessKeyId", request.GetAction())
		return
	}
	accessKeyId = *response.Response.Result.AccessKeyId
	return
}

func (me *APIGatewayService) EnableApiKey(ctx context.Context, accessKeyId string) (errRet error) {
	request := apigateway.NewEnableApiKeyRequest()
	request.AccessKeyId = &accessKeyId
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseAPIGatewayClient().EnableApiKey(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}
	if *response.Response.Result {
		return
	}
	return fmt.Errorf("enable api key fail")
}

func (me *APIGatewayService) DisableApiKey(ctx context.Context, accessKeyId string) (errRet error) {
	request := apigateway.NewDisableApiKeyRequest()
	request.AccessKeyId = &accessKeyId
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseAPIGatewayClient().DisableApiKey(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}
	if *response.Response.Result {
		return
	}
	return fmt.Errorf("disable api key fail")
}

func (me *APIGatewayService) DescribeApiKey(ctx context.Context,
	accessKeyId string) (apiKey *apigateway.ApiKey, has bool, errRet error) {
	apiKeySet, err := me.DescribeApiKeysStatus(ctx, "", accessKeyId)
	if err != nil {
		errRet = err
		return
	}
	if len(apiKeySet) == 0 {
		return
	}
	has = true
	apiKey = apiKeySet[0]
	return
}

func (me *APIGatewayService) DescribeApiKeysStatus(ctx context.Context, secretName, accessKeyId string) (apiKeySet []*apigateway.ApiKey, errRet error) {
	request := apigateway.NewDescribeApiKeysStatusRequest()
	if secretName != "" || accessKeyId != "" {
		request.Filters = make([]*apigateway.Filter, 0, 2)
		if secretName != "" {
			request.Filters = append(request.Filters, &apigateway.Filter{Name: helper.String("SecretName"),
				Values: []*string{
					&secretName,
				}})
		}
		if accessKeyId != "" {
			request.Filters = append(request.Filters, &apigateway.Filter{Name: helper.String("AccessKeyId"),
				Values: []*string{
					&accessKeyId,
				}})
		}
	}

	var limit int64 = 20
	var offset int64 = 0

	request.Limit = &limit
	request.Offset = &offset

	for {
		ratelimit.Check(request.GetAction())
		response, err := me.client.UseAPIGatewayClient().DescribeApiKeysStatus(request)
		if err != nil {
			errRet = err
			return
		}
		if response.Response.Result == nil {
			errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
			return
		}
		if len(response.Response.Result.ApiKeySet) > 0 {
			apiKeySet = append(apiKeySet, response.Response.Result.ApiKeySet...)
		}
		if len(response.Response.Result.ApiKeySet) < int(limit) {
			return
		}
		offset += limit
	}
}

func (me *APIGatewayService) DeleteApiKey(ctx context.Context, accessKeyId string) (errRet error) {
	request := apigateway.NewDeleteApiKeyRequest()
	request.AccessKeyId = &accessKeyId
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseAPIGatewayClient().DeleteApiKey(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}
	if *response.Response.Result {
		return
	}
	return fmt.Errorf("delete api key fail")
}

func (me *APIGatewayService) CreateUsagePlan(ctx context.Context,
	usagePlanName string,
	usagePlanDesc *string,
	maxRequestNum,
	maxRequestNumPreSec int64) (usagePlanId string, errRet error) {

	request := apigateway.NewCreateUsagePlanRequest()
	request.MaxRequestNum = &maxRequestNum
	request.MaxRequestNumPreSec = &maxRequestNumPreSec
	if usagePlanDesc != nil {
		request.UsagePlanDesc = usagePlanDesc
	}
	request.UsagePlanName = &usagePlanName

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseAPIGatewayClient().CreateUsagePlan(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}
	usagePlanId = *response.Response.Result.UsagePlanId
	return
}

func (me *APIGatewayService) DescribeUsagePlan(ctx context.Context, usagePlanId string) (info apigateway.UsagePlanInfo, has bool, errRet error) {

	request := apigateway.NewDescribeUsagePlanRequest()
	request.UsagePlanId = &usagePlanId

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseAPIGatewayClient().DescribeUsagePlan(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok && sdkErr.GetCode() == "ResourceNotFound.InvalidUsagePlan" {
			return
		}
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}
	has = true
	info = *response.Response.Result
	return
}

func (me *APIGatewayService) DeleteUsagePlan(ctx context.Context, usagePlanId string) (errRet error) {

	request := apigateway.NewDeleteUsagePlanRequest()
	request.UsagePlanId = &usagePlanId

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseAPIGatewayClient().DeleteUsagePlan(request)

	if err != nil {
		return err
	}
	if response.Response.Result == nil {
		return fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
	}

	if !*response.Response.Result {
		return fmt.Errorf("delete usage plan fail")
	}

	return
}

func (me *APIGatewayService) ModifyUsagePlan(ctx context.Context,
	usagePlanId string,
	usagePlanName string,
	usagePlanDesc *string,
	maxRequestNum,
	maxRequestNumPreSec int64)(errRet error){

	request := apigateway.NewModifyUsagePlanRequest()
	request.UsagePlanId = &usagePlanId

	ratelimit.Check(request.GetAction())
	request.UsagePlanName = &usagePlanName
	if usagePlanDesc!=nil{
		request.UsagePlanDesc = usagePlanDesc
	}
	request.MaxRequestNum = &maxRequestNum
	request.MaxRequestNumPreSec = &maxRequestNumPreSec

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseAPIGatewayClient().ModifyUsagePlan(request)
	if err != nil {
		errRet = err
		return
	}
	if response.Response.Result == nil {
		errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
		return
	}

	return nil
}

func (me *APIGatewayService) DescribeUsagePlanEnvironments(ctx context.Context,
	usagePlanId string,bindType string)(list []*apigateway. UsagePlanEnvironment,errRet error)  {

	request := apigateway.NewDescribeUsagePlanEnvironmentsRequest()
	request.UsagePlanId = &usagePlanId
	request.BindType = &bindType

	var limit int64 = 20
	var offset int64 = 0

	request.Limit = &limit
	request.Offset = &offset

	for {
		ratelimit.Check(request.GetAction())
		response, err := me.client.UseAPIGatewayClient().DescribeUsagePlanEnvironments(request)
		if err != nil {
			errRet = err
			return
		}
		if response.Response.Result == nil {
			errRet = fmt.Errorf("TencentCloud SDK %s return empty response", request.GetAction())
			return
		}
		if len(response.Response.Result.EnvironmentList) > 0 {
			list = append(list, response.Response.Result.EnvironmentList...)
		}
		if len(response.Response.Result.EnvironmentList) < int(limit) {
			return
		}
		offset += limit
	}
}








