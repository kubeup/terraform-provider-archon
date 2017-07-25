package kubernetes

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/builtin/providers/google"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"archon": testAccProvider,
		"google": google.Provider(),
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func TestProvider_configure(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()

	os.Setenv("KUBECONFIG", "test-fixtures/kube-config.yaml")
	os.Setenv("KUBE_CTX", "gcp")

	c, err := config.NewRawConfig(map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	rc := terraform.NewResourceConfig(c)
	p := Provider()
	err = p.Configure(rc)
	if err != nil {
		t.Fatal(err)
	}
}

func unsetEnv(t *testing.T) func() {
	e := getEnv()

	if err := os.Unsetenv("KUBECONFIG"); err != nil {
		t.Fatalf("Error unsetting env var KUBECONFIG: %s", err)
	}
	if err := os.Unsetenv("KUBE_CONFIG"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CONFIG: %s", err)
	}
	if err := os.Unsetenv("KUBE_CTX"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CTX: %s", err)
	}
	if err := os.Unsetenv("KUBE_CTX_AUTH_INFO"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CTX_AUTH_INFO: %s", err)
	}
	if err := os.Unsetenv("KUBE_CTX_CLUSTER"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CTX_CLUSTER: %s", err)
	}
	if err := os.Unsetenv("KUBE_HOST"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_HOST: %s", err)
	}
	if err := os.Unsetenv("KUBE_USER"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_USER: %s", err)
	}
	if err := os.Unsetenv("KUBE_PASSWORD"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_PASSWORD: %s", err)
	}
	if err := os.Unsetenv("KUBE_CLIENT_CERT_DATA"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CLIENT_CERT_DATA: %s", err)
	}
	if err := os.Unsetenv("KUBE_CLIENT_KEY_DATA"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CLIENT_KEY_DATA: %s", err)
	}
	if err := os.Unsetenv("KUBE_CLUSTER_CA_CERT_DATA"); err != nil {
		t.Fatalf("Error unsetting env var KUBE_CLUSTER_CA_CERT_DATA: %s", err)
	}

	return func() {
		if err := os.Setenv("KUBE_CONFIG", e.Config); err != nil {
			t.Fatalf("Error resetting env var KUBE_CONFIG: %s", err)
		}
		if err := os.Setenv("KUBECONFIG", e.Config); err != nil {
			t.Fatalf("Error resetting env var KUBECONFIG: %s", err)
		}
		if err := os.Setenv("KUBE_CTX", e.Config); err != nil {
			t.Fatalf("Error resetting env var KUBE_CTX: %s", err)
		}
		if err := os.Setenv("KUBE_CTX_AUTH_INFO", e.CtxAuthInfo); err != nil {
			t.Fatalf("Error resetting env var KUBE_CTX_AUTH_INFO: %s", err)
		}
		if err := os.Setenv("KUBE_CTX_CLUSTER", e.CtxCluster); err != nil {
			t.Fatalf("Error resetting env var KUBE_CTX_CLUSTER: %s", err)
		}
		if err := os.Setenv("KUBE_HOST", e.Host); err != nil {
			t.Fatalf("Error resetting env var KUBE_HOST: %s", err)
		}
		if err := os.Setenv("KUBE_USER", e.User); err != nil {
			t.Fatalf("Error resetting env var KUBE_USER: %s", err)
		}
		if err := os.Setenv("KUBE_PASSWORD", e.Password); err != nil {
			t.Fatalf("Error resetting env var KUBE_PASSWORD: %s", err)
		}
		if err := os.Setenv("KUBE_CLIENT_CERT_DATA", e.ClientCertData); err != nil {
			t.Fatalf("Error resetting env var KUBE_CLIENT_CERT_DATA: %s", err)
		}
		if err := os.Setenv("KUBE_CLIENT_KEY_DATA", e.ClientKeyData); err != nil {
			t.Fatalf("Error resetting env var KUBE_CLIENT_KEY_DATA: %s", err)
		}
		if err := os.Setenv("KUBE_CLUSTER_CA_CERT_DATA", e.ClusterCACertData); err != nil {
			t.Fatalf("Error resetting env var KUBE_CLUSTER_CA_CERT_DATA: %s", err)
		}
	}
}

func getEnv() *currentEnv {
	e := &currentEnv{
		Ctx:               os.Getenv("KUBE_CTX_CLUSTER"),
		CtxAuthInfo:       os.Getenv("KUBE_CTX_AUTH_INFO"),
		CtxCluster:        os.Getenv("KUBE_CTX_CLUSTER"),
		Host:              os.Getenv("KUBE_HOST"),
		User:              os.Getenv("KUBE_USER"),
		Password:          os.Getenv("KUBE_PASSWORD"),
		ClientCertData:    os.Getenv("KUBE_CLIENT_CERT_DATA"),
		ClientKeyData:     os.Getenv("KUBE_CLIENT_KEY_DATA"),
		ClusterCACertData: os.Getenv("KUBE_CLUSTER_CA_CERT_DATA"),
	}
	if cfg := os.Getenv("KUBE_CONFIG"); cfg != "" {
		e.Config = cfg
	}
	if cfg := os.Getenv("KUBECONFIG"); cfg != "" {
		e.Config = cfg
	}
	return e
}

func testAccPreCheck(t *testing.T) {
	hasFileCfg := (os.Getenv("KUBE_CTX_AUTH_INFO") != "" && os.Getenv("KUBE_CTX_CLUSTER") != "")
	hasStaticCfg := (os.Getenv("KUBE_HOST") != "" &&
		os.Getenv("KUBE_USER") != "" &&
		os.Getenv("KUBE_PASSWORD") != "" &&
		os.Getenv("KUBE_CLIENT_CERT_DATA") != "" &&
		os.Getenv("KUBE_CLIENT_KEY_DATA") != "" &&
		os.Getenv("KUBE_CLUSTER_CA_CERT_DATA") != "")

	if !hasFileCfg && !hasStaticCfg {
		t.Fatalf("File config (KUBE_CTX_AUTH_INFO and KUBE_CTX_CLUSTER) or static configuration"+
			" (%s) must be set for acceptance tests",
			strings.Join([]string{
				"KUBE_HOST",
				"KUBE_USER",
				"KUBE_PASSWORD",
				"KUBE_CLIENT_CERT_DATA",
				"KUBE_CLIENT_KEY_DATA",
				"KUBE_CLUSTER_CA_CERT_DATA",
			}, ", "))
	}
}

type currentEnv struct {
	Config            string
	Ctx               string
	CtxAuthInfo       string
	CtxCluster        string
	Host              string
	User              string
	Password          string
	ClientCertData    string
	ClientKeyData     string
	ClusterCACertData string
}

func testAccCheckMetaAnnotations(om *meta_v1.ObjectMeta, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 && len(om.Annotations) == 0 {
			return nil
		}
		if !reflect.DeepEqual(om.Annotations, expected) {
			return fmt.Errorf("%s annotations don't match.\nExpected: %q\nGiven: %q",
				om.Name, expected, om.Annotations)
		}
		return nil
	}
}

func testAccCheckMetaLabels(om *meta_v1.ObjectMeta, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 && len(om.Labels) == 0 {
			return nil
		}
		if !reflect.DeepEqual(om.Labels, expected) {
			return fmt.Errorf("%s labels don't match.\nExpected: %q\nGiven: %q",
				om.Name, expected, om.Labels)
		}
		return nil
	}
}
