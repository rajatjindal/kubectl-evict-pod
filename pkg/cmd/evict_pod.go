package cmd

import (
	"fmt"
	"os"

	"time"

	"github.com/rajatjindal/kubectl-evict-pod/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	withRetry     bool
	allPods       bool
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
	cmd.Flags().StringVarP(&o.labelSelector, "label-selector", "l", "", "label selector to evict pods with")
	cmd.Flags().StringVar(&o.labelSelector, "field-selector", "", "field selector to evict pods with")
	cmd.Flags().BoolVar(&o.withRetry, "retry", false, "retry eviction until it succeeds")
	cmd.Flags().BoolVar(&o.allPods, "all", false, "evict all pods in the namespace")
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
	hasPodNames := len(o.podNames) > 0
	hasSelectors := o.labelSelector != "" || o.fieldSelector != "" || o.allPods

	if hasPodNames && hasSelectors {
		return fmt.Errorf("pod name cannot be provided when a selector is specified")
	}

	if !hasPodNames && !hasSelectors {
		return fmt.Errorf("nothing given to select pods")
	}

	if o.allPods && (o.labelSelector != "" || o.fieldSelector != "") {
		return fmt.Errorf("--all cannot be used with label or field selectors")
	}

	return nil
}

// Run fetches the given secret manifest from the cluster, decodes the payload, opens an editor to make changes, and applies the modified manifest when done
func (o *EvictPodOptions) Run() error {
	if o.labelSelector != "" || o.fieldSelector != "" || o.allPods {
		options := v1.ListOptions{LabelSelector: o.labelSelector, FieldSelector: o.fieldSelector}
		var err error
		o.podNames, err = k8s.ListPods(o.kubeclient, o.namespace, options)
		if err != nil {
			return err
		}
	}

	return o.retryEachElement(o.podNames, func(podName string) error {
		err := k8s.Evict(o.kubeclient, podName, o.namespace)
		if err != nil {
			logrus.Warnf("pod %s in namespace %s evicting failed", podName, o.namespace)
		} else {
			logrus.Infof("pod %s in namespace %s evicted successfully", podName, o.namespace)
		}
		return err
	})
}

func (o *EvictPodOptions) retryEachElement(elements []string, fn func(string) error) error {
	if o.withRetry { // retry until every element passes
		var failed []string
		for {
			for _, podName := range elements {
				if err := fn(podName); err != nil {
					failed = append(failed, podName)
				}
			}
			if len(failed) == 0 {
				return nil
			} else {
				elements = failed
				failed = nil
				time.Sleep(5 * time.Second)
			}
		}
	} else { // just iterate once
		for _, element := range elements {
			if err := fn(element); err != nil {
				return err
			}
		}
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
