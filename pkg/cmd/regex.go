package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	%[1]s regex-match get pods "^nginx-"

	# list all services ending with "web" in namespace "foo"
	%[1]s regex-match get services "web$"
	
	# delete all configMaps in the "foo" namespace containing "app"
	%[1]s regex-match delete configMaps "app" -n foo
	`
	kubeFlags *genericclioptions.ConfigFlags

	allNamespaces bool
	autoYes       bool
)

func NewRegExCmd(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "regex-match",
		Short:        "Use RegEx to manage Kubernetes resources",
		Example:      fmt.Sprintf(RegexExample, "kubectl"),
		SilenceUsage: true,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl regex-match",
		},
	}
	kubeFlags = genericclioptions.NewConfigFlags(true)
	kubeFlags.AddFlags(cmd.PersistentFlags())

	// support --all-namespaces
	cmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If present, list across all namespaces")
	cmd.PersistentFlags().BoolVarP(&autoYes, "yes", "y", false, "Skip confirmation prompts and delete directly")

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
		matched := []struct {
			NS, Name string
		}{}

		for _, item := range list.Items {
			name := item.GetName()
			ns := item.GetNamespace()
			if re.MatchString(name) {
				matched = append(matched, struct {
					NS, Name string
				}{ns, name})
			}
		}

		if len(matched) == 0 {
			fmt.Fprintln(streams.Out, "No resources matched your pattern.")
			return nil
		}

		// Display matches
		fmt.Fprintf(streams.Out, "The following %s match your regex:\n", resource)
		for _, m := range matched {
			if m.NS != "" {
				fmt.Fprintf(streams.Out, "  %s/%s\n", m.NS, m.Name)
			} else {
				fmt.Fprintf(streams.Out, "  %s\n", m.Name)
			}
		}

		// Ask for confirmation once (unless --yes)
		if !autoYes {
			fmt.Fprintf(streams.Out, "\nDelete all %d resources? [y/N]: ", len(matched))
			var confirm string
			fmt.Fscanln(streams.In, &confirm)
			if strings.ToLower(confirm) != "y" {
				fmt.Fprintln(streams.Out, "Aborted.")
				return nil
			}
		}

		// Rebuild client for proper namespace scoping
		restCfg, err := kubeFlags.ToRESTConfig()
		if err != nil {
			return err
		}
		dynClient, err := dynamic.NewForConfig(restCfg)
		if err != nil {
			return err
		}
		mapper, err := kubeFlags.ToRESTMapper()
		if err != nil {
			return err
		}
		gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
		if err != nil {
			return fmt.Errorf("unknown resource %q: %w", resource, err)
		}

		// Delete all confirmed matches
		deleted, failed := 0, 0

		// Prepare the base resource interface (namespaceable)
		baseRI := dynClient.Resource(gvr)

		for _, m := range matched {
			var targetRI dynamic.ResourceInterface

			// For namespaced resources, re-scope
			if m.NS != "" {
				targetRI = baseRI.Namespace(m.NS)
			} else {
				targetRI = baseRI
			}

			if err := targetRI.Delete(context.Background(), m.Name, metav1.DeleteOptions{}); err != nil {
				fmt.Fprintf(streams.ErrOut, "Failed to delete %s/%s: %v\n", m.NS, m.Name, err)
				failed++
			} else {
				fmt.Fprintf(streams.Out, "Deleted %s/%s\n", m.NS, m.Name)
				deleted++
			}
		}

		fmt.Fprintf(streams.Out, "\n✅ %d deleted, ❌ %d failed.\n", deleted, failed)

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
	if gvkResource.Resource == "nodes" || gvkResource.Resource == "namespaces" || allNamespaces {
		// cluster-scoped
		ri = dynClient.Resource(gvkResource)
	} else {
		// namespaced
		ri = dynClient.Resource(gvkResource).Namespace(ns)
	}
	return ri, nil
}
