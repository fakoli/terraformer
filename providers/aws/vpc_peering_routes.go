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
	"github.com/hashicorp/terraform/helper/hashcode"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var routesAllowEmptyValues = []string{"tags."}

type VpcPeeringRoutesGenerator struct {
	AWSService
}

// Generate TerraformResources from AWS API,
// create terraform resource for each VPC Peering Connection
func (g *VpcPeeringRoutesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ec2.NewFromConfig(config)
	r := ec2.NewDescribeRouteTablesPaginator(svc, &ec2.DescribeRouteTablesInput{})
	for r.HasMorePages() {
		page, err := r.NextPage(context.TODO())
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.createRoutes(page)...)
	}
	return nil
}
func (g *VpcPeeringRoutesGenerator) createRoutes(routetables *ec2.DescribeRouteTablesOutput) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for _, table := range routetables.RouteTables {
		routeTableID := StringValue(table.RouteTableId)
		for _, route := range table.Routes {
			if route.VpcPeeringConnectionId == nil {
				continue
			}
			if route.DestinationCidrBlock == nil {
				continue
			}
			routeId := RouteCreateID(routeTableID, StringValue(route.DestinationCidrBlock))
			resources = append(resources, terraformutils.NewResource(
				routeId,
				routeId,
				"aws_route",
				"aws",
				map[string]string{
					"route_table_id":            routeTableID,
					"vpc_peering_connection_id": StringValue(route.VpcPeeringConnectionId),
					"destination_cidr_block":    StringValue(route.DestinationCidrBlock),
				},
				routesAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return resources
}

func RouteCreateID(routeTableID, destination string) string {
	return fmt.Sprintf("r-%s%d", routeTableID, hashcode.String(destination))
}

func RouteCreateName(routeTableName, id string) string {
	return fmt.Sprintf("%s%d", routeTableName, hashcode.String(id))
}
