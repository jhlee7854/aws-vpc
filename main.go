package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpc, err := createVpc(ctx)
		if err != nil {
			return err
		}

		publicSubnets, privateSubnets, err := SetDefaultSubnets(ctx, vpc)
		if err != nil {
			return err
		}

		err = SetDefaultInternetGateway(ctx, vpc, publicSubnets)
		if err != nil {
			return err
		}

		err = SetDefaultNATGateway(ctx, vpc, publicSubnets.Subnets[0], privateSubnets.Subnets[0])
		if err != nil {
			return err
		}

		err = SetDefaultNATGateway(ctx, vpc, publicSubnets.Subnets[1], privateSubnets.Subnets[1])
		if err != nil {
			return err
		}

		ctx.Export("vpcId", vpc.ID())
		subnetIds := []pulumi.IDOutput{}
		for _, subnet := range privateSubnets.Subnets {
			subnetIds = append(subnetIds, subnet.ID())
		}
		ctx.Export("subnetIds", pulumi.ToIDArrayOutput(subnetIds))

		return nil
	})
}

func Last(slice []string) string {
	return slice[len(slice)-1]
}
