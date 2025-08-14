package main

import (
	"fmt"
	"strings"

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

func createRouteTableForIGWAssociations(ctx *pulumi.Context, name string, routeTableId pulumi.StringInput, publicSubnetId pulumi.IDOutput, opts ...pulumi.ResourceOption) error {
	_, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
		RouteTableId: routeTableId,
		SubnetId:     publicSubnetId,
	}, opts...)
	if err != nil {
		return err
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

	for _, subnet := range publicSubnets.Subnets {
		subnet.URN().ToStringOutput().ApplyT(func(urn string) error {
			name := strings.ReplaceAll(Last(strings.Split(urn, "::")), "subnet", "rta")
			err = createRouteTableForIGWAssociations(ctx, name, rtForIgw.ID(), subnet.ID(), pulumi.Parent(rtForIgw), pulumi.DependsOn([]pulumi.Resource{subnet}))
			if err != nil {
				return err
			}
			return nil
		})
	}

	return nil
}
