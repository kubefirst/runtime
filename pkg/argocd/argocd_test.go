package argocd_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kubefirst/runtime/configs"
	"github.com/kubefirst/runtime/pkg/helpers"
	"github.com/spf13/viper"
)

// this is called when ArgoCD is up and running
func TestArgoCDLivenessIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config := configs.ReadConfig()
	err := helpers.SetupViper(config)
	if err != nil {
		t.Error(err)
	}

	var argoURL string
	if viper.GetString("cloud") == helpers.CloudK3d {
		argoURL = "http://localhost:8080"
	} else {
		argoURL = fmt.Sprintf("https://argocd.%s", viper.GetString("aws.hostedzonename"))
	}

	req, err := http.NewRequest(http.MethodGet, argoURL, nil)
	if err != nil {
		t.Error(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("wanted http status code 200, got %d", res.StatusCode)
	}
}

// this is called when Argo Workflow is up and running
func TestArgoWorkflowLivenessIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config := configs.ReadConfig()
	err := helpers.SetupViper(config)
	if err != nil {
		t.Error(err)
	}

	var argoURL string
	if viper.GetString("cloud") == helpers.CloudK3d {
		argoURL = "http://localhost:2746"
	} else {
		argoURL = fmt.Sprintf("https://argo.%s", viper.GetString("aws.hostedzonename"))
	}

	req, err := http.NewRequest(http.MethodGet, argoURL, nil)
	if err != nil {
		t.Error(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("wanted http status code 200, got %d", res.StatusCode)
	}
}