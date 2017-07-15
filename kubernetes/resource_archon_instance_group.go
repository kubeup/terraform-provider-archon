package kubernetes

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	pkgApi "k8s.io/apimachinery/pkg/types"
	archon "kubeup.com/archon/pkg/clientset"
	"kubeup.com/archon/pkg/cluster"
)

func resourceArchonInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceArchonInstanceGroupCreate,
		Read:   resourceArchonInstanceGroupRead,
		Exists: resourceArchonInstanceGroupExists,
		Update: resourceArchonInstanceGroupUpdate,
		Delete: resourceArchonInstanceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("instance_group", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Archon InstanceGroup spec",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"replicas": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"provision_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "DynamicOnly",
						},
						"selector": {
							Type:        schema.TypeList,
							Description: "A label query over volumes to consider for binding.",
							Optional:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"match_expressions": {
										Type:        schema.TypeList,
										Description: "A list of label selector requirements. The requirements are ANDed.",
										Optional:    true,
										ForceNew:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Description: "The label key that the selector applies to.",
													Optional:    true,
													ForceNew:    true,
												},
												"operator": {
													Type:        schema.TypeString,
													Description: "A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.",
													Optional:    true,
													ForceNew:    true,
												},
												"values": {
													Type:        schema.TypeSet,
													Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.",
													Optional:    true,
													ForceNew:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
											},
										},
									},
									"match_labels": {
										Type:        schema.TypeMap,
										Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
										Optional:    true,
										ForceNew:    true,
									},
								},
							},
						},
						"reserved_instance_selector": {
							Type:        schema.TypeList,
							Description: "A label query over volumes to consider for binding.",
							Optional:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"match_expressions": {
										Type:        schema.TypeList,
										Description: "A list of label selector requirements. The requirements are ANDed.",
										Optional:    true,
										ForceNew:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Description: "The label key that the selector applies to.",
													Optional:    true,
													ForceNew:    true,
												},
												"operator": {
													Type:        schema.TypeString,
													Description: "A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.",
													Optional:    true,
													ForceNew:    true,
												},
												"values": {
													Type:        schema.TypeSet,
													Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.",
													Optional:    true,
													ForceNew:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
											},
										},
									},
									"match_labels": {
										Type:        schema.TypeMap,
										Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
										Optional:    true,
										ForceNew:    true,
									},
								},
							},
						},
						"template": {
							Type:        schema.TypeList,
							Description: "Archon Instance spec",
							Required:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metadata": namespacedMetadataSchema("instance_template", true),
									"spec": {
										Type:        schema.TypeList,
										Description: "Archon Instance spec",
										Required:    true,
										ForceNew:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: instanceSpecFields(),
										},
									},
									"secrets": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metadata": namespacedMetadataSchema("secret", true),
												"data": {
													Type:        schema.TypeMap,
													Description: "A map of the secret data.",
													Optional:    true,
													Sensitive:   true,
												},
												"type": {
													Type:        schema.TypeString,
													Description: "Type of secret",
													Default:     "Opaque",
													Optional:    true,
													ForceNew:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceArchonInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	instanceGroup := cluster.InstanceGroup{
		ObjectMeta: metadata,
		Spec:       expandInstanceGroupSpec(d.Get("spec").([]interface{})),
	}
	log.Printf("[INFO] Creating new instance_group: %#v", instanceGroup)
	out, err := conn.Archon().InstanceGroups(metadata.Namespace).Create(&instanceGroup)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new instance_group: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonInstanceGroupRead(d, meta)
}

func resourceArchonInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Reading instance_group %s", name)
	instanceGroup, err := conn.Archon().InstanceGroups(namespace).Get(name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received instance_group: %#v", instanceGroup)
	err = d.Set("metadata", flattenMetadata(instanceGroup.ObjectMeta))
	if err != nil {
		return err
	}

	flattened := flattenInstanceGroupSpec(instanceGroup.Spec)
	log.Printf("[DEBUG] Flattened instance_group spec: %#v", flattened)
	err = d.Set("spec", flattened)
	if err != nil {
		return err
	}

	return nil
}

func resourceArchonInstanceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		diffOps := patchInstanceGroupSpec("spec.0.", "/spec/", d)
		ops = append(ops, diffOps...)
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating instance_group: %s", ops)
	out, err := conn.Archon().InstanceGroups(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated instance_group: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonInstanceGroupRead(d, meta)
}

func resourceArchonInstanceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Deleting instance_group: %#v", name)
	err := conn.Archon().InstanceGroups(namespace).Delete(name)
	if err != nil {
		return err
	}

	log.Printf("[INFO] InstanceGroup %s deleted", name)

	d.SetId("")
	return nil
}

func resourceArchonInstanceGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Checking instance_group %s", name)
	_, err := conn.Archon().InstanceGroups(namespace).Get(name)
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] InstanceGroup %s exists", name)
	return true, err
}
