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

		ctx.Export("vpcId", vpc.ID())

		return nil
	})
}
