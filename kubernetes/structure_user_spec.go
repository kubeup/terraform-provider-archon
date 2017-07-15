package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"kubeup.com/archon/pkg/cluster"
)

// Flatteners

func flattenUserSpec(in cluster.UserSpec) []interface{} {
	att := make(map[string]interface{})
	if in.Name != "" {
		att["name"] = in.Name
	}
	if in.PasswordHash != "" {
		att["password_hash"] = in.PasswordHash
	}
	if len(in.SSHAuthorizedKeys) > 0 {
		att["ssh_authorized_keys"] = newStringSet(schema.HashString, in.SSHAuthorizedKeys)
	}
	if in.Sudo != "" {
		att["sudo"] = in.Sudo
	}
	if in.Shell != "" {
		att["shell"] = in.Shell
	}
	return []interface{}{att}
}

// Expanders

func expandUserSpec(l []interface{}) cluster.UserSpec {
	if len(l) == 0 || l[0] == nil {
		return cluster.UserSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := cluster.UserSpec{}

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}
	if v, ok := in["password_hash"].(string); ok {
		obj.PasswordHash = v
	}
	if v, ok := in["external_ips"].(*schema.Set); ok && v.Len() > 0 {
		obj.SSHAuthorizedKeys = sliceOfString(v.List())
	}
	if v, ok := in["sudo"].(string); ok {
		obj.Sudo = v
	}
	if v, ok := in["shell"].(string); ok {
		obj.Shell = v
	}
	return obj
}

// Patch Ops

func patchUserSpec(keyPrefix, pathPrefix string, d *schema.ResourceData) PatchOperations {
	ops := make([]PatchOperation, 0, 0)
	if d.HasChange(keyPrefix + "name") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "name",
			Value: d.Get(keyPrefix + "name").(string),
		})
	}
	if d.HasChange(keyPrefix + "password_hash") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "passwordHash",
			Value: d.Get(keyPrefix + "password_hash").(string),
		})
	}
	if d.HasChange(keyPrefix + "ssh_authorized_keys") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "sshAuthorizedKeys",
			Value: d.Get(keyPrefix + "ssh_authorized_keys").(*schema.Set).List(),
		})
	}
	if d.HasChange(keyPrefix + "sudo") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "sudo",
			Value: d.Get(keyPrefix + "sudo").(string),
		})
	}
	if d.HasChange(keyPrefix + "shell") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "shell",
			Value: d.Get(keyPrefix + "shell").(string),
		})
	}
	return ops
}
