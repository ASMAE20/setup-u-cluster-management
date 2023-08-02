package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

var cmdArgs []string

func K8s_Tool(tool string, memory string, cpu string, profile string) {

	fmt.Println("Creation of your management cluster ...")
	_, err := exec.Command(tool, "start", "--memory", memory, "--cpus", cpu, "-p", profile).Output()
	if err != nil {
		fmt.Println("Error starting Minikube:", err)
		return
	}
	fmt.Println("your management cluster has been created successfully")

}

func Argo(kubeconfig string) {

	if kubeconfig != "" {
		cmdArgs = []string{"--kubeconfig", kubeconfig}
	}

	_, err := exec.Command("kubectl", cmdArgs[0], cmdArgs[1], "create", "namespace", "argocd").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("argocd namespace created")

	_, err2 := exec.Command("kubectl", cmdArgs[0], cmdArgs[1], "apply", "-n", "argocd", "-f", "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml").Output()
	if err2 != nil {
		fmt.Println("Error:", err2)
		return
	}
	fmt.Println("argo is installed")

	for {
		if PodsReady("argocd", kubeconfig) {
			break
		}
		fmt.Println("Waiting for argoCD-pods to be ready...")
		time.Sleep(60 * time.Second) // Sleep for seconds before checking again
	}
	fmt.Println("All pods in the argocd namespace are ready")
}

func CrossPlane(kubeconfig string) {

	_, err0 := exec.Command("helm", "repo", "add", "crossplane-stable", "https://charts.crossplane.io/stable").Output()
	if err0 != nil {
		fmt.Println("Error:", err0)
		return
	}
	_, err := exec.Command("helm", "repo", "update").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, err2 := exec.Command("kubectl", cmdArgs[0], cmdArgs[1], "create", "namespace", "crossplane-system").Output()
	if err2 != nil {
		fmt.Println("Error:", err2)
		return
	}
	fmt.Println("crossplane-system namespace created")
	_, err3 := exec.Command("helm", "upgrade", "--install", "crossplane", "--namespace", "crossplane-system", "crossplane-stable/crossplane", "--kubeconfig", kubeconfig).Output()
	if err3 != nil {
		fmt.Println("Error:", err3)
		return
	}
	fmt.Println("crossplane is installed")
	for {
		if PodsReady("crossplane-system", kubeconfig) {
			break
		}
		fmt.Println("Waiting for crossplane-pods to be ready...")
		time.Sleep(30 * time.Second) // Sleep for seconds before checking again
	}
	fmt.Println("All pods in crossplane-system namespace are ready")
}

func Providers(providers []string, kubeconfig string) {
	aroConfig := fmt.Sprintf(`
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
    name: my-cluster
    namespace: argocd
spec:
    project: default
    source:
        repoURL: https://github.com/younesELouafi/provider.git
        targetRevision: HEAD
        path: providers
        directory:
          include: '%s'
    destination:
        server: https://kubernetes.default.svc
        namespace: providers
    syncPolicy:
        automated:
            prune: true
            selfHeal: true
        syncOptions:
        - CreateNamespace=true`, ConvertToString(providers))

	cmd := exec.Command("kubectl", cmdArgs[0], cmdArgs[1], "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(aroConfig)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	time.Sleep(60 * time.Second)
	for {
		if ProvidersReady(kubeconfig) {
			break
		}
		fmt.Println("Waiting for providers to be ready...")
		time.Sleep(60 * time.Second) // Sleep for seconds before checking again
	}
	fmt.Println("Providers are ready")

}
func PodsReady(namespace string, kubeconfig string) bool {
	cmd := exec.Command("kubectl", "get", cmdArgs[0], cmdArgs[1], "pods", "-n", namespace, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	type PodStatus struct {
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	}

	type PodList struct {
		Items []PodStatus `json:"items"`
	}

	var pods PodList
	err = json.Unmarshal(output, &pods)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				return false
			}
		}
	}

	return true
}

func ProvidersReady(kubeconfig string) bool {
	cmd := exec.Command("kubectl", cmdArgs[0], cmdArgs[1], "get", "providers", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error here:", err)
		return false
	}

	type ProviderStatus struct {
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	}

	type ProviderList struct {
		Items []ProviderStatus `json:"items"`
	}

	var providers ProviderList
	err = json.Unmarshal(output, &providers)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	for _, provider := range providers.Items {
		for _, condition := range provider.Status.Conditions {
			if condition.Type == "Healthy" && condition.Status != "True" {
				return false
			}
		}
	}

	return true
}
func ConvertToString(providers []string) string {
	for i, str := range providers {
		providers[i] = str + ".yaml"
	}

	return "{" + strings.Join(providers, ",") + "}"
}

func askProv() []string {
	opts := []string{
		"helm",
		"k8s",
		"github",
		"aws",
		"scaleway",
		"azure",
		"googles",
	}

	choosedProviders := []string{}
	prompt := &survey.MultiSelect{
		Message: "Choose providers",
		Options: opts,
	}
	survey.AskOne(prompt, &choosedProviders)

	return choosedProviders
}