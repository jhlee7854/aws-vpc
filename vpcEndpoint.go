package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createVpcEndpointForS3(ctx *pulumi.Context, routeTableIds pulumi.StringArray, vpcId pulumi.IDOutput, opts ...pulumi.ResourceOption) (*ec2.VpcEndpoint, error) {
	vpceName := fmt.Sprintf("%s-vpce-s3", ctx.Stack())
	return ec2.NewVpcEndpoint(ctx, vpceName, &ec2.VpcEndpointArgs{
		Policy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": [
						"s3:GetObject",
						"s3:ListBucket"
					],
					"Resource": [
						"arn:aws:s3:::your-bucket-name",
						"arn:aws:s3:::your-bucket-name/*"
					]
				}
			]
		}`),
		RouteTableIds: routeTableIds,
		ServiceName:   pulumi.String("com.amazonaws.ap-northeast-2.s3"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(vpceName),
		},
		VpcEndpointType: pulumi.String("Gateway"),
		VpcId:           vpcId,
	}, opts...)
}
