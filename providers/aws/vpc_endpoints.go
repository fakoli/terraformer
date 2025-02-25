// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var endpointAllowEmptyValues = []string{"tags."}

type VpcEndpointGenerator struct {
	AWSService
}

// Generate TerraformResources from AWS API,
// create terraform resource for each VPC Endpoint Connection
func (g *VpcEndpointGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	if err := g.loadEndpoints(svc); err != nil {
		return err
	}
	if err := g.loadServices(svc); err != nil {
		return err
	}
	return nil
}

func (g *VpcEndpointGenerator) loadEndpoints(svc *ec2.Client) error {
	p := ec2.NewDescribeVpcEndpointsPaginator(svc, &ec2.DescribeVpcEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, vpcEndpoint := range page.VpcEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(vpcEndpoint.VpcEndpointId),
				StringValue(vpcEndpoint.VpcEndpointId),
				"aws_vpc_endpoint",
				"aws",
				endpointAllowEmptyValues,
			))
		}
	}

	return nil
}

// PostConvertHook for add policy json as heredoc
func (g *VpcEndpointGenerator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type == "aws_vpc_endpoint" {
			if val, ok := g.Resources[i].Item["policy"]; ok {
				policy := g.escapeAwsInterpolation(val.(string))
				g.Resources[i].Item["policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, policy)
			}
		}
	}
	return nil
}

func (g *VpcEndpointGenerator) loadServices(svc *ec2.Client) error {
	p := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(svc, &ec2.DescribeVpcEndpointServiceConfigurationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ServiceConfiguration := range page.ServiceConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(ServiceConfiguration.ServiceId),
				StringValue(ServiceConfiguration.ServiceId),
				"aws_vpc_endpoint_service",
				"aws",
				endpointAllowEmptyValues,
			))
		}
	}
	return nil
}
