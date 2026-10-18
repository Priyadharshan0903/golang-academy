"""Serve the static courses locally, choosing a nearby port if needed."""
import argparse
import errno
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--port", type=int, default=8080)
    parser.add_argument("--directory", required=True)
    parser.add_argument("--base", default="/")
    args = parser.parse_args()
    if not 1024 <= args.port <= 65535:
        parser.error("Choose a preview port between 1024 and 65535.")

    handler = partial(SimpleHTTPRequestHandler, directory=args.directory)
    for port in range(args.port, min(args.port + 20, 65536)):
        try:
            server = ThreadingHTTPServer(("127.0.0.1", port), handler)
            break
        except OSError as error:
            if error.errno != errno.EADDRINUSE:
                parser.exit(1, f"Cannot start local preview: {error}\n")
    else:
        parser.exit(1, "Preview ports are busy. Try npm run preview -- --port 9000\n")

    if port != args.port:
        print(f"Port {args.port} is already in use; using port {port}.", flush=True)
    base = "/" + args.base.strip("/") + "/" if args.base.strip("/") else "/"
    print(f"V1: http://127.0.0.1:{port}{base}v1/", flush=True)
    print(f"V2: http://127.0.0.1:{port}{base}v2/", flush=True)
    print("Press Ctrl+C to stop the preview.", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()


if __name__ == "__main__":
    main()
