package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/hashicorp/terraform/helper/hashcode"
)

const AwsRouteResourceType = "aws_route"

func initAwsRouteMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsRouteResourceType, func(res *resource.Resource) map[string]string {
		attributes := map[string]string{
			"route_table_id": *res.Attributes().GetString("route_table_id"),
		}
		if ipv4 := res.Attributes().GetString("destination_cidr_block"); ipv4 != nil && *ipv4 != "" {
			attributes["destination_cidr_block"] = *ipv4
		}
		if ipv6 := res.Attributes().GetString("destination_ipv6_cidr_block"); ipv6 != nil && *ipv6 != "" {
			attributes["destination_ipv6_cidr_block"] = *ipv6
		}
		return attributes
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsRouteResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})

		val.DeleteIfDefault("vpc_endpoint_id")
		val.DeleteIfDefault("local_gateway_id")
		val.DeleteIfDefault("destination_cidr_block")
		val.DeleteIfDefault("destination_ipv6_cidr_block")
		val.DeleteIfDefault("egress_only_gateway_id")
		val.DeleteIfDefault("nat_gateway_id")
		val.DeleteIfDefault("instance_id")
		val.DeleteIfDefault("network_interface_id")
		val.DeleteIfDefault("transit_gateway_id")
		val.DeleteIfDefault("vpc_peering_connection_id")
		val.DeleteIfDefault("destination_prefix_list_id")
		val.DeleteIfDefault("instance_owner_id")
		val.DeleteIfDefault("carrier_gateway_id")
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRouteResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if rtID := val.GetString("route_table_id"); rtID != nil && *rtID != "" {
			attrs["Table"] = *rtID
		}
		if ipv4 := val.GetString("destination_cidr_block"); ipv4 != nil && *ipv4 != "" {
			attrs["Destination"] = *ipv4
		}
		if ipv6 := val.GetString("destination_ipv6_cidr_block"); ipv6 != nil && *ipv6 != "" {
			attrs["Destination"] = *ipv6
		}
		if prefix := val.GetString("destination_prefix_list_id"); prefix != nil && *prefix != "" {
			attrs["Destination"] = *prefix
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsRouteResourceType, resource.FlagDeepMode)
}

func CalculateRouteID(tableId, CidrBlock, Ipv6CidrBlock, PrefixListId *string) string {
	if CidrBlock != nil && *CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*CidrBlock))
	}

	if Ipv6CidrBlock != nil && *Ipv6CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*Ipv6CidrBlock))
	}

	if PrefixListId != nil && *PrefixListId != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*PrefixListId))
	}

	return ""
}
