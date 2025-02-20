package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-std/sdk/go/std"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type subnets struct {
	pulumi.ResourceState
	Subnets    []*ec2.Subnet
	SubnetType string
}

func NewPublicSubnets(ctx *pulumi.Context, opts ...pulumi.ResourceOption) (*subnets, error) {
	return newSubnetsResource(ctx, true, opts...)
}

func NewPrivateSubents(ctx *pulumi.Context, opts ...pulumi.ResourceOption) (*subnets, error) {
	return newSubnetsResource(ctx, false, opts...)
}

func newSubnetsResource(ctx *pulumi.Context, isPublic bool, opts ...pulumi.ResourceOption) (*subnets, error) {
	var resource subnets

	if isPublic {
		resource.SubnetType = "public"
	} else {
		resource.SubnetType = "private"
	}

	resourceName := fmt.Sprintf("%s-subnets-%s", ctx.Stack(), resource.SubnetType)
	err := ctx.RegisterComponentResource("main:github.com/jhlee7854/aws-infra:subnets", resourceName, &resource, opts...)
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

type subnetArgs struct {
	VpcId            pulumi.StringInput
	VpcCidrBlock     pulumi.StringInput
	AvailabilityZone string
	Netnum           int
	Newbits          int
}

func (s *subnets) AddSubnet(ctx *pulumi.Context, args *subnetArgs) (*ec2.Subnet, error) {
	cidrBlock := std.CidrsubnetOutput(ctx, std.CidrsubnetOutputArgs{
		Input:   args.VpcCidrBlock,
		Netnum:  pulumi.Int(args.Netnum),
		Newbits: pulumi.Int(args.Newbits),
	}, nil)

	subnetName := fmt.Sprintf("%s-subnet-%s%d-%s", ctx.Stack(), s.SubnetType, len(s.Subnets)+1, args.AvailabilityZone)
	subnet, err := ec2.NewSubnet(ctx, subnetName, &ec2.SubnetArgs{
		AvailabilityZone:               pulumi.String(args.AvailabilityZone),
		CidrBlock:                      cidrBlock.Result(),
		PrivateDnsHostnameTypeOnLaunch: pulumi.String("ip-name"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(subnetName),
		},
		VpcId: args.VpcId,
	}, pulumi.Parent(s))
	if err != nil {
		return nil, err
	}

	s.Subnets = append(s.Subnets, subnet)

	return subnet, nil
}

func SetDefaultSubnets(ctx *pulumi.Context, vpc *ec2.Vpc) (*subnets, *subnets, error) {
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State:        pulumi.StringRef("available"),
		ExcludeNames: []string{"ap-northeast-2b", "ap-northeast-2d"},
	})
	if err != nil {
		return nil, nil, err
	}

	publicSubnets, err := NewPublicSubnets(ctx, pulumi.Parent(vpc))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		registerResourceMap := make(pulumi.Map)
		for i, subnet := range publicSubnets.Subnets {
			registerResourceMap[fmt.Sprintf("publicSubnet%d", i+1)] = subnet.ID()
		}
		ctx.RegisterResourceOutputs(publicSubnets, registerResourceMap)
	}()

	_, err = publicSubnets.AddSubnet(ctx, &subnetArgs{
		VpcId:            vpc.ID(),
		VpcCidrBlock:     vpc.CidrBlock,
		AvailabilityZone: azs.Names[0],
		Netnum:           0,
		Newbits:          4,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = publicSubnets.AddSubnet(ctx, &subnetArgs{
		VpcId:            vpc.ID(),
		VpcCidrBlock:     vpc.CidrBlock,
		AvailabilityZone: azs.Names[1],
		Netnum:           1,
		Newbits:          4,
	})
	if err != nil {
		return nil, nil, err
	}

	privateSubnets, err := NewPrivateSubents(ctx, pulumi.Parent(vpc))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		registerResourceMap := make(pulumi.Map)
		for i, subnet := range privateSubnets.Subnets {
			registerResourceMap[fmt.Sprintf("privateSubnet%d", i+1)] = subnet.ID()
		}
		ctx.RegisterResourceOutputs(privateSubnets, registerResourceMap)
	}()

	_, err = privateSubnets.AddSubnet(ctx, &subnetArgs{
		VpcId:            vpc.ID(),
		VpcCidrBlock:     vpc.CidrBlock,
		AvailabilityZone: azs.Names[0],
		Netnum:           2,
		Newbits:          4,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = privateSubnets.AddSubnet(ctx, &subnetArgs{
		VpcId:            vpc.ID(),
		VpcCidrBlock:     vpc.CidrBlock,
		AvailabilityZone: azs.Names[1],
		Netnum:           3,
		Newbits:          4,
	})
	if err != nil {
		return nil, nil, err
	}

	return publicSubnets, privateSubnets, nil
}
