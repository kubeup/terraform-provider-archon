package kubernetes

import (
	"kubeup.com/archon/pkg/cluster"
)

// Flatteners

func flattenInstanceSpec(in cluster.InstanceSpec) []interface{} {
	att := make(map[string]interface{})
	if in.OS != "" {
		att["os"] = in.OS
	}
	if in.Image != "" {
		att["image"] = in.Image
	}
	if in.InstanceType != "" {
		att["instance_type"] = in.InstanceType
	}
	if in.NetworkName != "" {
		att["network_name"] = in.NetworkName
	}
	if in.ReclaimPolicy != "" {
		att["reclaim_policy"] = in.ReclaimPolicy
	}
	if len(in.Files) > 0 {
		att["files"] = flattenFiles(in.Files)
	}
	if len(in.Secrets) > 0 {
		att["secrets"] = flattenArchonLocalObjectReferenceArray(in.Secrets)
	}
	if len(in.Configs) > 0 {
		att["configs"] = flattenConfigs(in.Configs)
	}
	if len(in.Users) > 0 {
		att["users"] = flattenArchonLocalObjectReferenceArray(in.Users)
	}
	if in.Hostname != "" {
		att["hostname"] = in.Hostname
	}
	if in.ReservedInstanceRef != nil {
		att["reserved_instance_ref"] = flattenArchonLocalObjectReferenceArray([]cluster.LocalObjectReference{*in.ReservedInstanceRef})
	}
	return []interface{}{att}
}

func flattenFiles(files []cluster.FileSpec) []interface{} {
	att := make([]interface{}, len(files))
	for i, v := range files {
		att[i] = flattenFileSpec(v)
	}
	return att
}

func flattenFileSpec(in cluster.FileSpec) map[string]interface{} {
	att := make(map[string]interface{})
	if in.Name != "" {
		att["name"] = in.Name
	}

	if in.Encoding != "" {
		att["encoding"] = in.Encoding
	}

	if in.Content != "" {
		att["content"] = in.Content
	}

	if in.Template != "" {
		att["template"] = in.Template
	}

	if in.UserID != 0 {
		att["user_id"] = in.UserID
	}

	if in.GroupID != 0 {
		att["group_id"] = in.GroupID
	}

	if in.Filesystem != "" {
		att["filesystem"] = in.Filesystem
	}

	if in.Path != "" {
		att["path"] = in.Path
	}

	if in.RawFilePermissions != "" {
		att["raw_file_permissions"] = in.RawFilePermissions
	}
	return att
}

func flattenArchonLocalObjectReferenceArray(in []cluster.LocalObjectReference) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		m := map[string]interface{}{}
		if v.Name != "" {
			m["name"] = v.Name
		}
		att[i] = m
	}
	return att
}

func flattenConfigs(in []cluster.ConfigSpec) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		m := map[string]interface{}{}
		if v.Name != "" {
			m["name"] = v.Name
		}
		if len(v.Data) > 0 {
			m["data"] = v.Data
		}
		att[i] = m
	}
	return att
}

// Expanders

func expandInstanceSpec(l []interface{}) cluster.InstanceSpec {
	if len(l) == 0 || l[0] == nil {
		return cluster.InstanceSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := cluster.InstanceSpec{}

	if v, ok := in["os"].(string); ok {
		obj.OS = v
	}
	if v, ok := in["image"].(string); ok {
		obj.Image = v
	}
	if v, ok := in["instance_type"].(string); ok {
		obj.InstanceType = v
	}
	if v, ok := in["network_name"].(string); ok {
		obj.NetworkName = v
	}
	if v, ok := in["files"].([]interface{}); ok {
		obj.Files = expandFiles(v)
	}
	if v, ok := in["secrets"].([]interface{}); ok {
		obj.Secrets = expandArchonLocalObjectReferenceArray(v)
	}
	if v, ok := in["configs"].([]interface{}); ok {
		obj.Configs = expandConfigs(v)
	}
	if v, ok := in["users"].([]interface{}); ok {
		obj.Users = expandArchonLocalObjectReferenceArray(v)
	}
	if v, ok := in["hostname"].(string); ok {
		obj.Hostname = v
	}
	if v, ok := in["reserved_instance_ref"].([]interface{}); ok {
		c := expandArchonLocalObjectReferenceArray(v)
		if len(c) > 0 {
			obj.ReservedInstanceRef = &c[0]
		}
	}
	return obj
}

func expandFiles(in []interface{}) []cluster.FileSpec {
	if len(in) == 0 {
		return []cluster.FileSpec{}
	}
	files := make([]cluster.FileSpec, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		if v, ok := p["name"]; ok {
			files[i].Name = v.(string)
		}
		if v, ok := p["encoding"]; ok {
			files[i].Encoding = v.(string)
		}
		if v, ok := p["content"]; ok {
			files[i].Content = v.(string)
		}
		if v, ok := p["template"]; ok {
			files[i].Content = v.(string)
		}
		if v, ok := p["owner"]; ok {
			files[i].Content = v.(string)
		}
		if v, ok := p["user_id"]; ok {
			files[i].UserID = v.(int)
		}
		if v, ok := p["group_id"]; ok {
			files[i].GroupID = v.(int)
		}
		if v, ok := p["filesystem"]; ok {
			files[i].Filesystem = v.(string)
		}
		if v, ok := p["path"]; ok {
			files[i].Path = v.(string)
		}
		if v, ok := p["raw_file_permissions"]; ok {
			files[i].Path = v.(string)
		}
	}
	return files
}

func expandConfigs(in []interface{}) []cluster.ConfigSpec {
	if len(in) == 0 {
		return []cluster.ConfigSpec{}
	}
	configs := make([]cluster.ConfigSpec, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		if v, ok := p["name"]; ok {
			configs[i].Name = v.(string)
		}
		if v, ok := p["data"]; ok {
			configs[i].Data = v.(map[string]string)
		}
	}
	return configs
}

func expandArchonLocalObjectReferenceArray(in []interface{}) []cluster.LocalObjectReference {
	att := []cluster.LocalObjectReference{}
	if len(in) < 1 {
		return att
	}
	att = make([]cluster.LocalObjectReference, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		if name, ok := p["name"]; ok {
			att[i].Name = name.(string)
		}
	}
	return att
}
