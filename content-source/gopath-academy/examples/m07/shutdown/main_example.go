package shutdown

// [start main]

// A complete main function for a service would look like this:
//
//	func main() {
//		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//		defer stop()
//
//		ln, err := net.Listen("tcp", ":8080")
//		if err != nil {
//			log.Fatal(err)
//		}
//		if err := serve(ctx, ln, newRouter()); err != nil {
//			log.Fatal(err)
//		}
//		log.Println("stopped cleanly")
//	}
//
// Kubernetes sends SIGTERM, waits terminationGracePeriodSeconds
// (30 seconds by default), then sends SIGKILL. The shutdown deadline in
// serve must be shorter than that grace period.

// [end main]
