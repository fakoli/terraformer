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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var peeringAllowEmptyValues = []string{"tags."}

type VpcPeeringConnectionGenerator struct {
	AWSService
}

// Generate TerraformResources from AWS API,
// create terraform resource for each VPC Peering Connection
func (g *VpcPeeringConnectionGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	p := ec2.NewDescribeVpcPeeringConnectionsPaginator(svc, &ec2.DescribeVpcPeeringConnectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.createVpcConnections(page)...)
	}

	return nil
}

func (g *VpcPeeringConnectionGenerator) createVpcConnections(peerings *ec2.DescribeVpcPeeringConnectionsOutput) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for _, peering := range peerings.VpcPeeringConnections {
		peeringID := StringValue(peering.VpcPeeringConnectionId)

		if StringValue(peering.AccepterVpcInfo.Region) == StringValue(peering.RequesterVpcInfo.Region) {
			resources = append(resources, terraformutils.NewResource(
				peeringID,
				peeringID,
				"aws_vpc_peering_connection",
				"aws",
				map[string]string{
					"peer_owner_id": StringValue(peering.AccepterVpcInfo.OwnerId),
					"auto_accept":   "true",
					"peer_vpc_id":   StringValue(peering.AccepterVpcInfo.VpcId),
					"vpc_id":        StringValue(peering.RequesterVpcInfo.VpcId),
				},
				peeringAllowEmptyValues,
				map[string]interface{}{},
			))
			continue
		}

		if StringValue(peering.RequesterVpcInfo.Region) == g.GetArgs()["region"] && StringValue(peering.AccepterVpcInfo.Region) != g.GetArgs()["region"] {

			resources = append(resources, terraformutils.NewResource(
				peeringID,
				peeringID,
				"aws_vpc_peering_connection",
				"aws",
				map[string]string{
					"peer_owner_id": StringValue(peering.AccepterVpcInfo.OwnerId),
					"auto_accept":   "false",
					"peer_region":   StringValue(peering.AccepterVpcInfo.Region),
					"peer_vpc_id":   StringValue(peering.AccepterVpcInfo.VpcId),
					"vpc_id":        StringValue(peering.RequesterVpcInfo.VpcId),
				},
				peeringAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return resources
}
