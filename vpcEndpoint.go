package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createVpcEndpointForS3(ctx *pulumi.Context, routeTableIds pulumi.StringArray, vpcId pulumi.IDOutput, opts ...pulumi.ResourceOption) (*ec2.VpcEndpoint, error) {
	vpceName := fmt.Sprintf("%s-vpce-s3", ctx.Stack())
	return ec2.NewVpcEndpoint(ctx, vpceName, &ec2.VpcEndpointArgs{
		Policy:        pulumi.String("{\"Statement\":[{\"Action\":\"*\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}],\"Version\":\"2008-10-17\"}"),
		RouteTableIds: routeTableIds,
		ServiceName:   pulumi.String("com.amazonaws.ap-northeast-2.s3"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(vpceName),
		},
		VpcEndpointType: pulumi.String("Gateway"),
		VpcId:           vpcId,
	}, opts...)
}
