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

func resourceArchonInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceArchonInstanceCreate,
		Read:   resourceArchonInstanceRead,
		Exists: resourceArchonInstanceExists,
		Update: resourceArchonInstanceUpdate,
		Delete: resourceArchonInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("instance", true),
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
		},
	}
}

func resourceArchonInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	instance := cluster.Instance{
		ObjectMeta: metadata,
		Spec:       expandInstanceSpec(d.Get("spec").([]interface{})),
	}
	log.Printf("[INFO] Creating new instance: %#v", instance)
	out, err := conn.Archon().Instances(metadata.Namespace).Create(&instance)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new instance: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonInstanceRead(d, meta)
}

func resourceArchonInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Reading instance %s", name)
	instance, err := conn.Archon().Instances(namespace).Get(name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received instance: %#v", instance)
	err = d.Set("metadata", flattenMetadata(instance.ObjectMeta))
	if err != nil {
		return err
	}

	flattened := flattenInstanceSpec(instance.Spec)
	log.Printf("[DEBUG] Flattened instance spec: %#v", flattened)
	err = d.Set("spec", flattened)
	if err != nil {
		return err
	}

	return nil
}

func resourceArchonInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating instance: %s", ops)
	out, err := conn.Archon().Instances(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated instance: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonInstanceRead(d, meta)
}

func resourceArchonInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Deleting instance: %#v", name)
	err := conn.Archon().Instances(namespace).Delete(name)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Instance %s deleted", name)

	d.SetId("")
	return nil
}

func resourceArchonInstanceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Checking instance %s", name)
	_, err := conn.Archon().Instances(namespace).Get(name)
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] Instance %s exists", name)
	return true, err
}
