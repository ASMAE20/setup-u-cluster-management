package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/manifoldco/promptui"

	setup "github.com/ASMAE20/setup-u-cluster-management/pkg"

)

var kubeconfig string

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "This command helps you create your custom cluster setup.",
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := cmd.Flags().Lookup("kubeconfig").Value.String()

		if kubeconfig == "" {
			promptLocalK8s := promptui.Select{
				Label: "Kubernetes local tool",
				Items: []string{"minikube", "kind", "k3d"},
			}

			_, Local_K8s, errLocalK8s := promptLocalK8s.Run()
			if errLocalK8s != nil {
				fmt.Printf("Prompt failed %v\n", errLocalK8s)
				return
			}

			promptProfile := promptui.Prompt{
				Label: "Name of your management-cluster",
			}

			profile, errProfile := promptProfile.Run()
			if errProfile != nil {
				fmt.Printf("Prompt failed %v\n", errProfile)
				return
			}

			promptMemory := promptui.Prompt{
				Label: "Memory",
			}

			memory, errMemory := promptMemory.Run()
			if errMemory != nil {
				fmt.Printf("Prompt failed %v\n", errMemory)
				return
			}

			promptCpus := promptui.Prompt{
				Label: "CPUs",
			}

			cpu, errCpus := promptCpus.Run()
			if errCpus != nil {
				fmt.Printf("Prompt failed %v\n", errCpus)
				return
			}
		    setup.K8s_Tool(Local_K8s, memory, cpu, profile)
		    setup.askProv()
		    
		} else {
			setup.askProv()
			setup.CrossPlane(kubeconfig)
			setup.Argo(kubeconfig)
			setup.Providers(setup.askProv(), kubeconfig)

		}
	
},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	// Define the kubeconfig flag
	setupCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
}
