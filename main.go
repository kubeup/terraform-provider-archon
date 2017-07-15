package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/kubeup/terraform-provider-archon/kubernetes"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: kubernetes.Provider})
}
