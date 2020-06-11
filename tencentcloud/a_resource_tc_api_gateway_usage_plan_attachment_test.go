package tencentcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	apigateway "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/apigateway/v20180808"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAPIGatewayUsagePlanAttachmentResourceName = "tencentcloud_api_gateway_usage_plan_attachment"
var testAPIGatewayUsagePlanAttachmentResourceKey = testAPIGatewayUsagePlanAttachmentResourceName + ".attach"

func TestAccTencentCloudAPIGateWayUsagePlanAttachmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAPIGatewayUsagePlanAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayUsagePlanAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIGatewayUsagePlanAttachmentExists(testAPIGatewayUsagePlanAttachmentResourceKey),
					resource.TestCheckResourceAttrSet(testAPIGatewayUsagePlanAttachmentResourceKey, "usage_plan_id"),
					resource.TestCheckResourceAttrSet(testAPIGatewayUsagePlanAttachmentResourceKey, "service_id"),
					resource.TestCheckResourceAttr(testAPIGatewayUsagePlanAttachmentResourceKey, "environment","test"),
					resource.TestCheckResourceAttr(testAPIGatewayUsagePlanAttachmentResourceKey, "bind_type","SERVICE"),
				),
			},
			{
				ResourceName:      testAPIGatewayUsagePlanAttachmentResourceKey,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAPIGatewayUsagePlanAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != testAPIGatewayUsagePlanAttachmentResourceName {
			continue
		}
		logId := getLogId(contextNil)
		ctx := context.WithValue(context.TODO(), logIdKey, logId)

		var idMap = make(map[string]string, 5)
		if err := json.Unmarshal([]byte(rs.Primary.ID), &idMap); err != nil {
			return fmt.Errorf("id is broken,%s", err.Error())
		}
		var (
			usagePlanId = idMap["usage_plan_id"]
			serviceId   = idMap["service_id"]
			environment = idMap["environment"]
			bindType    = idMap["bind_type"]
			apiId       = idMap["api_id"]

			outErr error
			has           bool
		)
		if usagePlanId == "" || serviceId == "" || environment == "" || bindType == "" {
			return fmt.Errorf("id is broken")
		}
		if bindType == API_GATEWAY_TYPE_API && apiId == "" {
			return fmt.Errorf("id is broken")
		}

		service := APIGatewayService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}

		_, has, outErr= service.DescribeUsagePlan(ctx, usagePlanId)
		if outErr != nil {
			_, has, outErr = service.DescribeUsagePlan(ctx, usagePlanId)
		}
		if outErr != nil {
			return outErr
		}
		if !has {
			return nil
		}

		_, has, outErr = service.DescribeService(ctx, serviceId)
		if outErr != nil {
			_, has, outErr = service.DescribeService(ctx, serviceId)
		}
		if outErr != nil {
			return outErr
		}
		if !has {
			return nil
		}

		var plans []*apigateway.ApiUsagePlan

		if bindType == API_GATEWAY_TYPE_API {
			plans, outErr = service.DescribeApiUsagePlan(ctx, serviceId)
			if outErr!=nil{
				plans, outErr = service.DescribeApiUsagePlan(ctx, serviceId)
			}
			if outErr!=nil{
				return  outErr
			}
		} else {
			plans, outErr = service.DescribeServiceUsagePlan(ctx, serviceId)
			if outErr!=nil{
				plans, outErr = service.DescribeServiceUsagePlan(ctx, serviceId)
			}
			if outErr!=nil{
				return  outErr
			}
		}

		for _, plan := range plans {
			if *plan.UsagePlanId == usagePlanId && *plan.Environment == environment {
				if bindType == API_GATEWAY_TYPE_API {
					if plan.ApiId != nil && *plan.ApiId == apiId {
						return fmt.Errorf("attachment  %s still exist on server",rs.Primary.ID)
					}
				} else {
					return fmt.Errorf("attachment  %s still exist on server",rs.Primary.ID)
				}
			}
		}

		return nil
	}
	return nil
}

func testAccCheckAPIGatewayUsagePlanAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource %s is not found", n)
		}
		logId := getLogId(contextNil)
		ctx := context.WithValue(context.TODO(), logIdKey, logId)

		var idMap = make(map[string]string, 5)
		if err := json.Unmarshal([]byte(rs.Primary.ID), &idMap); err != nil {
			return fmt.Errorf("id is broken,%s", err.Error())
		}
		var (
			usagePlanId = idMap["usage_plan_id"]
			serviceId   = idMap["service_id"]
			environment = idMap["environment"]
			bindType    = idMap["bind_type"]
			apiId       = idMap["api_id"]

			outErr error
			has           bool
		)
		if usagePlanId == "" || serviceId == "" || environment == "" || bindType == "" {
			return fmt.Errorf("id is broken")
		}
		if bindType == API_GATEWAY_TYPE_API && apiId == "" {
			return fmt.Errorf("id is broken")
		}

		service := APIGatewayService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}

		_, has, outErr= service.DescribeUsagePlan(ctx, usagePlanId)
		if outErr != nil {
			_, has, outErr = service.DescribeUsagePlan(ctx, usagePlanId)
		}
		if outErr != nil {
			return outErr
		}
		if !has {
			return fmt.Errorf("usage plan %s not exsit on server",usagePlanId)
		}

		_, has, outErr = service.DescribeService(ctx, serviceId)
		if outErr != nil {
			_, has, outErr = service.DescribeService(ctx, serviceId)
		}
		if outErr != nil {
			return outErr
		}
		if !has {
			return fmt.Errorf("service %s not exsit on server",serviceId)
		}

		var plans []*apigateway.ApiUsagePlan

		if bindType == API_GATEWAY_TYPE_API {
			plans, outErr = service.DescribeApiUsagePlan(ctx, serviceId)
			if outErr!=nil{
				plans, outErr = service.DescribeApiUsagePlan(ctx, serviceId)
			}
			if outErr!=nil{
				return  outErr
			}
		} else {
			plans, outErr = service.DescribeServiceUsagePlan(ctx, serviceId)
			if outErr!=nil{
				plans, outErr = service.DescribeServiceUsagePlan(ctx, serviceId)
			}
			if outErr!=nil{
				return  outErr
			}
		}
		for _, plan := range plans {
			if *plan.UsagePlanId == usagePlanId && *plan.Environment == environment {
				if bindType == API_GATEWAY_TYPE_API {
					if plan.ApiId != nil && *plan.ApiId == apiId {
						return nil
					}
				} else {
					return  nil
				}
			}
		}
		return fmt.Errorf("attachment  %s not exist on server",rs.Primary.ID)
	}
}

const testAccAPIGatewayUsagePlanAttachment = `
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
`
