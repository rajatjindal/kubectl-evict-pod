package cmd

import (
	"fmt"
	"os"

	"github.com/rajatjindal/kubectl-evict-pod/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Version is set during build time
var Version = "unknown"

type EvictPodOptions struct {
	configFlags *genericclioptions.ConfigFlags
	ioStreams   genericclioptions.IOStreams

	podNames      []string
	namespace     string
	kubeclient    kubernetes.Interface
	printVersion  bool
	labelSelector string
	fieldSelector string
}

// NewEvictPodOptions provides an instance of EvictPodOptions with default values
func NewEvictPodOptions(streams genericclioptions.IOStreams) *EvictPodOptions {
	return &EvictPodOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		ioStreams:   streams,
	}
}

// NewEvictCmd provides a cobra command wrapping EvictPodOptions
func NewEvictCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewEvictPodOptions(streams)

	cmd := &cobra.Command{
		Use:          "evict-pod [flags]",
		Short:        "evict selected pods",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if o.printVersion {
				fmt.Println(Version)
				os.Exit(0)
			}

			if err := o.Complete(args); err != nil {
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
	cmd.Flags().StringVarP(&o.labelSelector, "labelSelector", "l", "", "labelSelector to evict pods with")
	cmd.Flags().StringVar(&o.labelSelector, "fieldSelector", "", "fieldSelector to evict pods with")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all config options
func (o *EvictPodOptions) Complete(args []string) error {
	o.podNames = args

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

// Validate ensures that config options are correct
func (o *EvictPodOptions) Validate() error {
	if len(o.podNames) > 0 && (o.labelSelector != "" || o.fieldSelector != "") {
		return fmt.Errorf("pod name cannot be provided when a selector is specified")
	}

	if len(o.podNames) == 0 && o.labelSelector == "" && o.fieldSelector == "" {
		return fmt.Errorf("nothing given to select pods")
	}

	return nil
}

// Run fetches the given secret manifest from the cluster, decodes the payload, opens an editor to make changes, and applies the modified manifest when done
func (o *EvictPodOptions) Run() error {
	var err error

	if o.labelSelector != "" || o.fieldSelector != "" {
		options := v1.ListOptions{LabelSelector: o.labelSelector, FieldSelector: o.fieldSelector}
		o.podNames, err = k8s.ListPods(o.kubeclient, o.namespace, options)
		if err != nil {
			return err
		}
	}

	for _, podName := range o.podNames {
		err := k8s.Evict(o.kubeclient, podName, o.namespace)
		if err != nil {
			return err
		}

		logrus.Infof("pod %s in namespace %s evicted successfully", podName, o.namespace)
	}

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
