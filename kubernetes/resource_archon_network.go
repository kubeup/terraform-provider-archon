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

func resourceArchonNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceArchonNetworkCreate,
		Read:   resourceArchonNetworkRead,
		Exists: resourceArchonNetworkExists,
		Update: resourceArchonNetworkUpdate,
		Delete: resourceArchonNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("network", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Archon Network spec",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceArchonNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	network := cluster.Network{
		ObjectMeta: metadata,
		Spec:       expandNetworkSpec(d.Get("spec").([]interface{})),
	}
	log.Printf("[INFO] Creating new network: %#v", network)
	out, err := conn.Archon().Networks(metadata.Namespace).Create(&network)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new network: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonNetworkRead(d, meta)
}

func resourceArchonNetworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Reading network %s", name)
	network, err := conn.Archon().Networks(namespace).Get(name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received network: %#v", network)
	err = d.Set("metadata", flattenMetadata(network.ObjectMeta))
	if err != nil {
		return err
	}

	flattened := flattenNetworkSpec(network.Spec)
	log.Printf("[DEBUG] Flattened network spec: %#v", flattened)
	err = d.Set("spec", flattened)
	if err != nil {
		return err
	}

	return nil
}

func resourceArchonNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		diffOps := patchNetworkSpec("spec.0.", "/spec/", d)
		ops = append(ops, diffOps...)
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating network: %s", ops)
	out, err := conn.Archon().Networks(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated network: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonNetworkRead(d, meta)
}

func resourceArchonNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Deleting network: %#v", name)
	err := conn.Archon().Networks(namespace).Delete(name)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Network %s deleted", name)

	d.SetId("")
	return nil
}

func resourceArchonNetworkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Checking network %s", name)
	_, err := conn.Archon().Networks(namespace).Get(name)
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] Network %s exists", name)
	return true, err
}
