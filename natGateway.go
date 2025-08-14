package main

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createEip(ctx *pulumi.Context, availabilityZone string, opts ...pulumi.ResourceOption) (*ec2.Eip, error) {
	cfg := config.New(ctx, "aws")
	region := cfg.Require("region")

	eipName := fmt.Sprintf("%s-eip-%s", ctx.Stack(), availabilityZone)
	return ec2.NewEip(ctx, eipName, &ec2.EipArgs{
		Domain:             pulumi.String("vpc"),
		NetworkBorderGroup: pulumi.String(region),
		PublicIpv4Pool:     pulumi.String("amazon"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(eipName),
		},
	}, opts...)
}

func createNatGateway(ctx *pulumi.Context, name string, eipId pulumi.IDOutput, publicSubnetId pulumi.IDOutput, opts ...pulumi.ResourceOption) (*ec2.NatGateway, error) {
	return ec2.NewNatGateway(ctx, name, &ec2.NatGatewayArgs{
		AllocationId:     eipId,
		ConnectivityType: pulumi.String("public"),
		SubnetId:         publicSubnetId,
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	}, opts...)
}

func createRouteTableForNGW(ctx *pulumi.Context, name string, vpcId pulumi.StringInput, ngwId pulumi.StringInput, opts ...pulumi.ResourceOption) (*ec2.RouteTable, error) {
	return ec2.NewRouteTable(ctx, name, &ec2.RouteTableArgs{
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock:    pulumi.String("0.0.0.0/0"),
				NatGatewayId: ngwId,
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
		VpcId: vpcId,
	}, opts...)
}

func createRouteTableForNGWAssociation(ctx *pulumi.Context, name string, routeTableId pulumi.StringInput, privateSubnetId pulumi.IDOutput, opts ...pulumi.ResourceOption) error {
	_, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
		RouteTableId: routeTableId,
		SubnetId:     privateSubnetId,
	}, opts...)
	if err != nil {
		return err
	}
	return nil
}

// 지정한 public subnet에 NAT Gateway를 만들고 지정한 private subnet들을 이용해 route table과 route table association을 생성한다.
// NAT Gateway를 만드는 과정 중 생성한 route table ID를 반환한다.
func SetDefaultNATGateway(ctx *pulumi.Context, vpc *ec2.Vpc, publicSubnet *ec2.Subnet, privateSubnets ...*ec2.Subnet) pulumi.IDOutput {
	return pulumi.All(publicSubnet.URN().ToStringOutput(), publicSubnet.AvailabilityZone).ApplyT(func(args []interface{}) (pulumi.IDOutput, error) {
		urn := args[0].(string)
		az := args[1].(string)

		eip, err := createEip(ctx, az, pulumi.Parent(vpc))
		if err != nil {
			return pulumi.IDOutput{}, err
		}

		ngwName := strings.ReplaceAll(Last(strings.Split(urn, "::")), "subnet", "nat")
		ngw, err := createNatGateway(ctx, ngwName, eip.ID(), publicSubnet.ID(), pulumi.Parent(vpc))
		if err != nil {
			return pulumi.IDOutput{}, err
		}

		rtForNgwID := privateSubnets[0].URN().ToStringOutput().ApplyT(func(urn string) (pulumi.IDOutput, error) {
			rtbName := strings.ReplaceAll(Last(strings.Split(urn, "::")), "subnet", "rtb")
			rtForNgw, err := createRouteTableForNGW(ctx, rtbName, vpc.ID(), ngw.ID(), pulumi.Parent(ngw))
			if err != nil {
				return pulumi.IDOutput{}, err
			}

			for _, privateSubnet := range privateSubnets {
				privateSubnet.URN().ToStringOutput().ApplyT(func(urn string) error {
					rtaName := strings.ReplaceAll(Last(strings.Split(urn, "::")), "subnet", "rta")
					err = createRouteTableForNGWAssociation(ctx, rtaName, rtForNgw.ID(), privateSubnet.ID(), pulumi.Parent(rtForNgw), pulumi.DependsOn([]pulumi.Resource{privateSubnet}))
					if err != nil {
						return err
					}

					return nil
				})
			}

			return rtForNgw.ID(), nil
		}).(pulumi.IDOutput)

		return rtForNgwID, nil
	}).(pulumi.IDOutput)
}
