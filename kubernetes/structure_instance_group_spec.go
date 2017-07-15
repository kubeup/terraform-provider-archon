package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api/v1"
	"kubeup.com/archon/pkg/cluster"
)

// Flatteners

func flattenInstanceGroupSpec(in cluster.InstanceGroupSpec) []interface{} {
	att := make(map[string]interface{})
	att["replicas"] = in.Replicas
	if in.ProvisionPolicy != "" {
		att["provision_policy"] = in.ProvisionPolicy
	}
	att["selector"] = flattenLabelSelector(in.Selector)
	att["reserved_instance_selector"] = flattenLabelSelector(in.ReservedInstanceSelector)
	att["template"] = flattenInstanceGroupTemplate(in.Template)

	return []interface{}{att}
}

func flattenInstanceGroupTemplate(in cluster.InstanceTemplateSpec) []interface{} {
	att := make(map[string]interface{})
	att["metadata"] = flattenMetadata(in.ObjectMeta)
	att["spec"] = flattenInstanceSpec(in.Spec)
	att["secrets"] = flattenSecrets(in.Secrets)
	return []interface{}{att}
}

func flattenSecrets(secrets []v1.Secret) []interface{} {
	att := make([]interface{}, len(secrets))
	for i, v := range secrets {
		att[i] = flattenSecret(v)
	}
	return att
}

func flattenSecret(in v1.Secret) map[string]interface{} {
	att := make(map[string]interface{})
	att["metadata"] = flattenMetadata(in.ObjectMeta)
	att["data"] = byteMapToStringMap(in.Data)
	att["type"] = in.Type
	return att
}

// Expanders

func expandInstanceGroupSpec(l []interface{}) cluster.InstanceGroupSpec {
	if len(l) == 0 || l[0] == nil {
		return cluster.InstanceGroupSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := cluster.InstanceGroupSpec{}

	if v, ok := in["replicas"].(int); ok {
		obj.Replicas = int32(v)
	}
	if v, ok := in["provision_policy"].(cluster.InstanceGroupProvisionPolicy); ok {
		obj.ProvisionPolicy = v
	}
	if v, ok := in["selector"].([]interface{}); ok {
		obj.Selector = expandLabelSelector(v)
	}
	if v, ok := in["reserved_instance_selector"].([]interface{}); ok {
		obj.ReservedInstanceSelector = expandLabelSelector(v)
	}
	if v, ok := in["template"].([]interface{}); ok {
		obj.Template = expandInstanceTemplateSpec(v)
	}
	return obj
}

func expandInstanceTemplateSpec(l []interface{}) cluster.InstanceTemplateSpec {
	if len(l) == 0 || l[0] == nil {
		return cluster.InstanceTemplateSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := cluster.InstanceTemplateSpec{}

	if v, ok := in["metadata"].([]interface{}); ok {
		obj.ObjectMeta = expandMetadata(v)
	}

	if v, ok := in["spec"].([]interface{}); ok {
		obj.Spec = expandInstanceSpec(v)
	}

	if v, ok := in["secrets"].([]interface{}); ok {
		obj.Secrets = expandSecrets(v)
	}

	return obj
}

func expandSecrets(in []interface{}) []v1.Secret {
	if len(in) == 0 {
		return []v1.Secret{}
	}
	secrets := make([]v1.Secret, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		if v, ok := p["metadata"].([]interface{}); ok {
			secrets[i].ObjectMeta = expandMetadata(v)
		}
		if v, ok := p["data"].(map[string]interface{}); ok {
			secrets[i].StringData = expandStringMap(v)
		}
		if v, ok := p["type"].(v1.SecretType); ok {
			secrets[i].Type = v
		}
	}
	return secrets
}

// Patch Ops

func patchInstanceGroupSpec(keyPrefix, pathPrefix string, d *schema.ResourceData) PatchOperations {
	ops := make([]PatchOperation, 0, 0)
	if d.HasChange(keyPrefix + "replicas") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "replicas",
			Value: d.Get(keyPrefix + "replicas").(int),
		})
	}
	return ops
}
