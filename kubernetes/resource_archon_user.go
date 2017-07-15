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

func resourceArchonUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceArchonUserCreate,
		Read:   resourceArchonUserRead,
		Exists: resourceArchonUserExists,
		Update: resourceArchonUserUpdate,
		Delete: resourceArchonUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("user", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Archon User spec",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"password_hash": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssh_authorized_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"sudo": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"shell": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceArchonUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	user := cluster.User{
		ObjectMeta: metadata,
		Spec:       expandUserSpec(d.Get("spec").([]interface{})),
	}
	log.Printf("[INFO] Creating new user: %#v", user)
	out, err := conn.Archon().Users(metadata.Namespace).Create(&user)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new user: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonUserRead(d, meta)
}

func resourceArchonUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Reading user %s", name)
	user, err := conn.Archon().Users(namespace).Get(name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received user: %#v", user)
	err = d.Set("metadata", flattenMetadata(user.ObjectMeta))
	if err != nil {
		return err
	}

	flattened := flattenUserSpec(user.Spec)
	log.Printf("[DEBUG] Flattened user spec: %#v", flattened)
	err = d.Set("spec", flattened)
	if err != nil {
		return err
	}

	return nil
}

func resourceArchonUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		diffOps := patchUserSpec("spec.0.", "/spec/", d)
		ops = append(ops, diffOps...)
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating user: %s", ops)
	out, err := conn.Archon().Users(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated user: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceArchonUserRead(d, meta)
}

func resourceArchonUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Deleting user: %#v", name)
	err := conn.Archon().Users(namespace).Delete(name)
	if err != nil {
		return err
	}

	log.Printf("[INFO] User %s deleted", name)

	d.SetId("")
	return nil
}

func resourceArchonUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*archon.Clientset)

	namespace, name := idParts(d.Id())
	log.Printf("[INFO] Checking user %s", name)
	_, err := conn.Archon().Users(namespace).Get(name)
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] User %s exists", name)
	return true, err
}
