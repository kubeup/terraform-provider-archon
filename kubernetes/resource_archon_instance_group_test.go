package kubernetes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	archon "kubeup.com/archon/pkg/clientset"
	"kubeup.com/archon/pkg/cluster"
)

func TestAccArchonInstanceGroup_basic(t *testing.T) {
	var conf cluster.InstanceGroup
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "archon_instancegroup.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckArchonInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchonInstanceGroupConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchonInstanceGroupExists("archon_instancegroup.test", &conf),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.TestAnnotationTwo", "two"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.%", "3"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.TestLabelTwo", "two"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelTwo": "two", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "spec.0.replicas", "2"),
				),
			},
			{
				Config: testAccArchonInstanceGroupConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchonInstanceGroupExists("archon_instancegroup.test", &conf),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.annotations.Different", "1234"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "Different": "1234"}),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.%", "2"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("archon_instancegroup.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("archon_instancegroup.test", "spec.0.replicas", "2"),
				),
			},
		},
	})
}

func TestAccArchonInstanceGroup_importBasic(t *testing.T) {
	resourceName := "archon_instancegroup.test"
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArchonInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchonInstanceGroupConfig_basic(name),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckArchonInstanceGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*archon.Clientset)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "archon_instancegroup" {
			continue
		}
		namespace, name := idParts(rs.Primary.ID)
		resp, err := conn.Archon().InstanceGroups(namespace).Get(name)
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("InstanceGroup still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckArchonInstanceGroupExists(n string, obj *cluster.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*archon.Clientset)
		namespace, name := idParts(rs.Primary.ID)
		out, err := conn.Archon().InstanceGroups(namespace).Get(name)
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccArchonInstanceGroupConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "archon_instancegroup" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			TestAnnotationTwo = "two"
		}
		labels {
			TestLabelOne = "one"
			TestLabelTwo = "two"
			TestLabelThree = "three"
		}
		name = "%s"
	}
	spec {
		replicas = 2
		selector {
			match_labels {
				app = "test"
			}
		}
		template {
			metadata {
				labels {
					app = "test"
				}
			}
			spec {
				image = "first"
				os = "second"
				network_name = "${archon_network.test.metadata.0.name}"
			}
		}
	}
}

resource "archon_network" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			TestAnnotationTwo = "two"
		}
		labels {
			TestLabelOne = "one"
			TestLabelTwo = "two"
			TestLabelThree = "three"
		}
		name = "tf-acc-network"
	}
	spec {
		region = "first"
		zone = "second"
		subnet = "10.0.0.0/24"
	}
}`, name)
}

func testAccArchonInstanceGroupConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "archon_instancegroup" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			Different = "1234"
		}
		labels {
			TestLabelOne = "one"
			TestLabelThree = "three"
		}
		name = "%s"
	}
	spec {
		replicas = 2
		selector {
			match_labels {
				app = "test"
			}
		}
		template {
			metadata {
				labels {
					app = "test"
				}
			}
			spec {
				image = "first"
				os = "second"
				network_name = "${archon_network.test.metadata.0.name}"
			}
		}
	}
}

resource "archon_network" "test" {
	metadata {
		annotations {
			TestAnnotationOne = "one"
			TestAnnotationTwo = "two"
		}
		labels {
			TestLabelOne = "one"
			TestLabelTwo = "two"
			TestLabelThree = "three"
		}
		name = "tf-acc-network"
	}
	spec {
		region = "first"
		zone = "second"
		subnet = "10.0.0.0/24"
	}
}`, name)
}
