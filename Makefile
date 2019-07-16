
SOURCES = $(shell echo *.go)
DEBUG_ASSETS := -debug=true
VERSION = $(shell grep '^const Version' version.go | sed -E 's/.*"([^"]+)"$$/\1/')

BUILD_OPTS = -ldflags="-s -w"
DIST := dist

all:
	@echo alertist $(VERSION)
	@echo $(SOURCES)

# for test
alertist: $(SOURCES) cmd/alertist/main.go
	go build -tags=linux $(BUILD_OPTS) -o $@ cmd/alertist/main.go

build:
	gox $(BUILD_OPTS) -osarch "linux/amd64 darwin/amd64 windows/amd64 windows/386" -output "$(DIST)/$(VERSION)/alertist-$(VERSION)_{{.OS}}_{{.Arch}}/alertist"

package: build
	-@mkdir $(DIST)/$(VERSION)/pkg 2>/dev/null
	@cd $(DIST)/$(VERSION) && for pkg in alertist-$(VERSION)_*; do \
	  echo $${pkg}; \
	  zip pkg/$${pkg}.zip $${pkg}/*; \
	done

release:
	$(MAKE) package
	@env ghr -u hirose31 --replace $(VERSION) $(DIST)/$(VERSION)/pkg/

clean:
	$(RM) *~ alertist
