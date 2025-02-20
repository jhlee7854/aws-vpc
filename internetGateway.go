package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createInternetGateway(ctx *pulumi.Context, vpcId pulumi.StringInput, opts ...pulumi.ResourceOption) (*ec2.InternetGateway, error) {
	igwName := fmt.Sprintf("%s-igw", ctx.Stack())
	return ec2.NewInternetGateway(ctx, igwName, &ec2.InternetGatewayArgs{
		Tags: pulumi.StringMap{
			"Name": pulumi.String(igwName),
		},
		VpcId: vpcId,
	}, opts...)
}

func createRouteTableForIGW(ctx *pulumi.Context, vpcId pulumi.StringInput, igwId pulumi.StringInput, opts ...pulumi.ResourceOption) (*ec2.RouteTable, error) {
	routeTableName := fmt.Sprintf("%s-rtb-pulbic", ctx.Stack())
	return ec2.NewRouteTable(ctx, routeTableName, &ec2.RouteTableArgs{
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: igwId,
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(routeTableName),
		},
		VpcId: vpcId,
	}, opts...)
}

func createRouteTableForIGWAssociations(ctx *pulumi.Context, routeTableId pulumi.StringInput, publicSubnets *subnets, opts ...pulumi.ResourceOption) error {
	for i, subnet := range publicSubnets.Subnets {
		subnet.AvailabilityZone.ApplyT(func(v string) error {
			rtaName := fmt.Sprintf("%s-subnet-public%d-%s", ctx.Stack(), i+1, v)
			_, err := ec2.NewRouteTableAssociation(ctx, rtaName, &ec2.RouteTableAssociationArgs{
				RouteTableId: routeTableId,
				SubnetId:     subnet.ID(),
			}, opts...)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}

func SetDefaultInternetGateway(ctx *pulumi.Context, vpc *ec2.Vpc, publicSubnets *subnets) error {
	igw, err := createInternetGateway(ctx, vpc.ID(), pulumi.Parent(vpc))
	if err != nil {
		return err
	}

	rtForIgw, err := createRouteTableForIGW(ctx, vpc.ID(), igw.ID(), pulumi.Parent(igw))
	if err != nil {
		return err
	}

	err = createRouteTableForIGWAssociations(ctx, rtForIgw.ID(), publicSubnets, pulumi.Parent(rtForIgw), pulumi.DependsOn([]pulumi.Resource{publicSubnets}))
	if err != nil {
		return err
	}

	return nil
}
