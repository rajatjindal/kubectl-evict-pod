package cmd

import (
	"fmt"
	"os"

	"github.com/rajatjindal/kubectl-evict-pod/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

//Version is set during build time
var Version = "unknown"

//EvictPodOptions is struct for modify secret
type EvictPodOptions struct {
	configFlags *genericclioptions.ConfigFlags
	iostreams   genericclioptions.IOStreams

	podNames     []string
	namespace    string
	kubeclient   kubernetes.Interface
	printVersion bool
	label        string
}

// NewEvictPodOptions provides an instance of EvictPodOptions with default values
func NewEvictPodOptions(streams genericclioptions.IOStreams) *EvictPodOptions {
	return &EvictPodOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		iostreams:   streams,
	}
}

// NewCmdModifySecret provides a cobra command wrapping EvictPodOptions
func NewCmdModifySecret(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewEvictPodOptions(streams)

	cmd := &cobra.Command{
		Use:          "evict-pod [flags]",
		Short:        "evict the pod selected by name or labels",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if o.printVersion {
				fmt.Println(Version)
				os.Exit(0)
			}

			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&o.printVersion, "version", false, "prints version of plugin")
	cmd.Flags().StringVar(&o.label, "label", "", "specify a label to evict pods with")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all information required for updating the current context
func (o *EvictPodOptions) Complete(cmd *cobra.Command, args []string) error {

	if len(args) > 0 {

		o.podNames = append(o.podNames, args[0])
	}

	if len(args) > 1 {
		o.podNames = args
	}

	config, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.kubeclient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	o.namespace = getNamespace(o.configFlags)
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *EvictPodOptions) Validate() error {

	if len(o.podNames) > 0 && o.label != "" {
		return fmt.Errorf("pod name cannot be provided when a selector is specified")
	}

	return nil
}

// Run fetches the given secret manifest from the cluster, decodes the payload, opens an editor to make changes, and applies the modified manifest when done
func (o *EvictPodOptions) Run() error {
	var err error

	if o.label != "" {

		o.podNames, err = k8s.PodsFromLabel(o.kubeclient, o.label, o.namespace)
		if err != nil {
			return err
		}
	}

	for _, podName := range o.podNames {
		err := k8s.Evict(o.kubeclient, podName, o.namespace)
		if err != nil {
			return err
		}
	}

	logrus.Infof("pods %q in namespace %s evicted successfully", o.podNames, o.namespace)
	return nil
}

// getNamespace takes a set of kubectl flag values and returns the namespace we should be operating in
func getNamespace(flags *genericclioptions.ConfigFlags) string {
	namespace, _, err := flags.ToRawKubeConfigLoader().Namespace()
	if err != nil || len(namespace) == 0 {
		namespace = "default"
	}
	return namespace
}
