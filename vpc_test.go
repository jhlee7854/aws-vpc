package main

import (
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// https://www.pulumi.com/docs/reference/pkg/nodejs/pulumi/pulumi/interfaces/runtime.Mocks.html
// https://github.com/pulumi/pulumi/blob/sdk/v3.150.0/sdk/go/pulumi/mocks.go#L34

type mocks int

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

// https://github.com/pulumi/pulumi/issues/4472#issuecomment-731012097
func withMocksWithConfig(project, stack string, config map[string]string, mocks pulumi.MockResourceMonitor) pulumi.RunOption {
	return func(info *pulumi.RunInfo) {
		info.Project, info.Stack, info.Config, info.Mocks = project, stack, config, mocks
	}
}

func TestCreateVpc(t *testing.T) {
	config := map[string]string{
		"vpc:cidr": "10.0.0.0/16",
	}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		vpc, err := createVpc(ctx)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)

		// check 1: Tag에 Name 키가 있어야 한다.
		pulumi.All(vpc.URN(), vpc.Tags).ApplyT(func(all []interface{}) error {
			urn := all[0].(pulumi.URN)
			tags := all[1].(map[string]string)

			assert.Containsf(t, tags, "Name", "missing a Name tag on VPC %v", urn)
			wg.Done()
			return nil
		})

		// check 2: EnableDnsHostnames 속성은 true 이다.
		pulumi.All(vpc.URN(), vpc.EnableDnsHostnames).ApplyT(func(all []interface{}) error {
			urn := all[0].(pulumi.URN)
			enableDnsHostnames := all[1].(bool)

			assert.Truef(t, enableDnsHostnames, "EnabelDnsHostnames property must have true on VPC %v", urn)
			wg.Done()
			return nil
		})

		wg.Wait()

		return nil
	}, withMocksWithConfig("project", "stack", config, mocks(0)))
	assert.NoError(t, err)
}
