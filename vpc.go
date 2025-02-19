package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createVpc(ctx *pulumi.Context) (*ec2.Vpc, error) {
	cfg := config.New(ctx, "vpc")
	vpcCidr := cfg.Require("cidr")

	vpcName := fmt.Sprintf("%s-vpc", ctx.Stack())
	return ec2.NewVpc(ctx, vpcName, &ec2.VpcArgs{
		CidrBlock:          pulumi.String(vpcCidr),
		EnableDnsHostnames: pulumi.Bool(true),
		InstanceTenancy:    pulumi.String("default"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(vpcName),
		},
	})
}
