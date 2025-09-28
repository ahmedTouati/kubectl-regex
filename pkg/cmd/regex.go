package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/dynamic"
)

var (
	RegexExample = `
	# list all pods starting with "nginx-" in current context's namespace
	%[1]s regex get pods "^nginx-"

	# list all services ending with "web" in namespace "foo"
	%[1]s regex get services "web$"
	
	# delete all configMaps in the "foo" namespace containing "app"
	%[1]s regex delete configMaps "app" -n foo
	`
	kubeFlags *genericclioptions.ConfigFlags
)

func NewRegExCmd(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "regex",
		Short:        "Use RegEx to manage Kubernetes resources",
		Example:      fmt.Sprintf(RegexExample, "kubectl"),
		SilenceUsage: true,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl regex",
		},
	}
	kubeFlags = genericclioptions.NewConfigFlags(true)
	kubeFlags.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(NewGetCmd(streams))
	cmd.AddCommand(NewDeleteCmd(streams))
	return cmd
}

func NewGetCmd(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <resource> [pattern]",
		Short: "Get Kubernetes resources matching RegEx",
		Args:  ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd(streams, args, "get")
		},
	}
	return cmd
}

func NewDeleteCmd(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <resource> [pattern]",
		Short: "Delete Kubernetes resources matching RegEx",
		Args:  ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd(streams, args, "delete")
		},
	}
	return cmd
}

func ValidateArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("resource type must be specified")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}
	return nil
}

func runCmd(streams genericiooptions.IOStreams, args []string, operation string) error {

	var pattern string
	if len(args) > 1 {
		pattern = args[1]
	}
	resource := args[0]

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	// Build client
	ri, err := BuildResourceInterface(resource)
	if err != nil {
		return err
	}

	list, err := ri.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Filter by regex
	switch operation {
	case "get":
		for _, item := range list.Items {
			name := item.GetName()
			if re.MatchString(name) {
				fmt.Println(name)
			}
		}
	case "delete":
		for _, item := range list.Items {
			name := item.GetName()
			if re.MatchString(name) {
				if err := ri.Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
					fmt.Fprintf(streams.ErrOut, "Failed to delete %s: %v\n", name, err)
				} else {
					fmt.Fprintf(streams.Out, "Deleted %s\n", name)
				}
			}
		}
	default:
		return fmt.Errorf("unknown operation %q", operation)
	}

	return nil
}

func BuildResourceInterface(resource string) (dynamic.ResourceInterface, error) {
	// Build client
	restCfg, err := kubeFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	// Find GVR (GroupVersionResource) for this resource
	mapper, err := kubeFlags.ToRESTMapper()
	if err != nil {
		return nil, err
	}
	gvkResource, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return nil, fmt.Errorf("unknown resource %q: %w", resource, err)
	}

	// Determine namespace
	ns, _, err := kubeFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	// Fetch list
	var ri dynamic.ResourceInterface
	if gvkResource.Resource == "nodes" || gvkResource.Resource == "namespaces" {
		// cluster-scoped
		ri = dynClient.Resource(gvkResource)
	} else {
		// namespaced
		ri = dynClient.Resource(gvkResource).Namespace(ns)
	}
	return ri, nil
}
