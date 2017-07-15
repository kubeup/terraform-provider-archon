package kubernetes

import "github.com/hashicorp/terraform/helper/schema"

func instanceSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"os": {
			Type:     schema.TypeString,
			Required: true,
			Computed: false,
		},
		"image": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"instance_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"network_name": {
			Type:     schema.TypeString,
			Required: true,
			Computed: false,
		},
		"reclaim_policy": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"files": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"encoding": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"content": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"template": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"owner": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"user_id": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"group_id": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"filesystem": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"path": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"raw_file_permissions": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"secrets": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     localObjectReferenceSchema(),
		},
		"configs": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"data": {
						Type:         schema.TypeMap,
						Optional:     true,
						ValidateFunc: validateAnnotations,
					},
				},
			},
		},
		"users": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     localObjectReferenceSchema(),
		},
		"hostname": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"reserved_instance_ref": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     localObjectReferenceSchema(),
		},
	}
}

func localObjectReferenceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
