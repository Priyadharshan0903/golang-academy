# Common tasks for working on the course.
#
#   make check     everything CI runs: style, vet, tests, site build
#   make test      test examples, projects, the server and exercise solutions
#                  (examples/thirdparty, projects/shortener, projects/jobqueue and
#                  projects/shortlink-operator are
#                  separate modules; database tests need TEST_DATABASE_URL and
#                  TEST_REDIS_ADDR, see examples/thirdparty/README.md)
#   make serve     run the site locally with live reload
#   make image     build the container image

.PHONY: check style vet test site serve image

check: style vet test site

style:
	./scripts/check-style.sh

vet:
	go vet ./...
	cd exercises && go vet ./... && go vet -tags solution ./...
	cd examples/thirdparty && go mod tidy && go vet ./...
	cd projects/shortener && go mod tidy && go vet ./...
	cd projects/jobqueue && go mod tidy && go vet ./...
	cd projects/shortlink-operator && go mod tidy && go vet ./...

test:
	go test ./...
	cd exercises && go test -tags solution ./...
	cd examples/thirdparty && go test ./...
	cd projects/shortener && go test ./...
	cd projects/jobqueue && go test ./...
	cd projects/shortlink-operator && go test ./...
	./scripts/check-mutants.sh

site:
	cd site && hugo --gc --minify

serve:
	cd site && hugo server --disableFastRender

image:
	docker build -f deploy/Dockerfile -t gopath-academy:dev .
