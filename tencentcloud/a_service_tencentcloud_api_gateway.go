package tencentcloud

import (
	"context"
	"fmt"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
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

func (me *APIGatewayService) DescribeApiKey(ctx context.Context, accessKeyId string) (apiKey *apigateway.ApiKey, has bool, errRet error) {
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
			request.Filters = append(request.Filters, &apigateway.Filter{Name: helper.String("SecretName"), Values: []*string{
				&secretName,
			}})
		}
		if accessKeyId != "" {
			request.Filters = append(request.Filters, &apigateway.Filter{Name: helper.String("AccessKeyId"), Values: []*string{
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
