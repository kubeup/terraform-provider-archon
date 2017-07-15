package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"kubeup.com/archon/pkg/cluster"
)

// Flatteners

func flattenNetworkSpec(in cluster.NetworkSpec) []interface{} {
	att := make(map[string]interface{})
	if in.Region != "" {
		att["region"] = in.Region
	}
	if in.Zone != "" {
		att["zone"] = in.Zone
	}
	if in.Subnet != "" {
		att["subnet"] = in.Subnet
	}
	return []interface{}{att}
}

// Expanders

func expandNetworkSpec(l []interface{}) cluster.NetworkSpec {
	if len(l) == 0 || l[0] == nil {
		return cluster.NetworkSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := cluster.NetworkSpec{}

	if v, ok := in["region"].(string); ok {
		obj.Region = v
	}
	if v, ok := in["zone"].(string); ok {
		obj.Zone = v
	}
	if v, ok := in["subnet"].(string); ok {
		obj.Subnet = v
	}
	return obj
}

// Patch Ops

func patchNetworkSpec(keyPrefix, pathPrefix string, d *schema.ResourceData) PatchOperations {
	ops := make([]PatchOperation, 0, 0)
	if d.HasChange(keyPrefix + "region") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "region",
			Value: d.Get(keyPrefix + "region").(string),
		})
	}
	if d.HasChange(keyPrefix + "zone") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "zone",
			Value: d.Get(keyPrefix + "zone").(string),
		})
	}
	if d.HasChange(keyPrefix + "subnet") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "subnet",
			Value: d.Get(keyPrefix + "subnet").(string),
		})
	}
	return ops
}
