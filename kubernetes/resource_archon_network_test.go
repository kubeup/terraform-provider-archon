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

func TestAccArchonNetwork_basic(t *testing.T) {
	var conf cluster.Network
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "archon_network.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckArchonNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchonNetworkConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchonNetworkExists("archon_network.test", &conf),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.TestAnnotationTwo", "two"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.%", "3"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.TestLabelTwo", "two"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelTwo": "two", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.region", "first"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.zone", "second"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.subnet", "10.0.0.0/24"),
				),
			},
			{
				Config: testAccArchonNetworkConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckArchonNetworkExists("archon_network.test", &conf),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.annotations.Different", "1234"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "Different": "1234"}),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.%", "2"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("archon_network.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("archon_network.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.region", "first"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.zone", "second"),
					resource.TestCheckResourceAttr("archon_network.test", "spec.0.subnet", "10.0.0.0/24"),
				),
			},
		},
	})
}

func TestAccArchonNetwork_importBasic(t *testing.T) {
	resourceName := "archon_network.test"
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckArchonNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchonNetworkConfig_basic(name),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckArchonNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*archon.Clientset)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "archon_network" {
			continue
		}
		namespace, name := idParts(rs.Primary.ID)
		resp, err := conn.Archon().Networks(namespace).Get(name)
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("Network still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckArchonNetworkExists(n string, obj *cluster.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*archon.Clientset)
		namespace, name := idParts(rs.Primary.ID)
		out, err := conn.Archon().Networks(namespace).Get(name)
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccArchonNetworkConfig_basic(name string) string {
	return fmt.Sprintf(`
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
		name = "%s"
	}
	spec {
		region = "first"
		zone = "second"
		subnet = "10.0.0.0/24"
	}
}`, name)
}

func testAccArchonNetworkConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "archon_network" "test" {
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
		region = "first"
		zone = "second"
		subnet = "10.0.0.0/24"
	}
}`, name)
}
