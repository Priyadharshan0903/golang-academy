// Command podlist prints the pods of a namespace, like a small
// `kubectl get pods`, using client-go (lesson 15.3):
//
//	go run ./k8s/podlist -n shortener
//	go run ./k8s/podlist -n shortener -watch
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/examples/thirdparty/k8s/kube"
)

func main() {
	ns := flag.String("n", "default", "namespace")
	selector := flag.String("l", "", "label selector")
	watch := flag.Bool("watch", false, "print pod events until interrupted")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	cs, err := kube.NewClientset()
	if err != nil {
		fmt.Fprintln(os.Stderr, "podlist:", err)
		os.Exit(1)
	}

	if *watch {
		events := make(chan kube.PodEvent)
		go func() {
			for e := range events {
				fmt.Printf("%-8s %-40s %s\n", e.Type, e.Name, e.Phase)
			}
		}()
		if err := kube.WatchPods(ctx, cs, *ns, events); err != nil {
			fmt.Fprintln(os.Stderr, "podlist:", err)
			os.Exit(1)
		}
		return
	}

	pods, err := kube.Pods(ctx, cs, *ns, *selector, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, "podlist:", err)
		os.Exit(1)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE")
	for _, p := range pods {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", p.Name, p.Ready, p.Phase, p.Restarts, p.Age)
	}
	tw.Flush()
}
