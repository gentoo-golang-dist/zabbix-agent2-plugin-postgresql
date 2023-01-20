PACKAGE=zabbix-agent2-plugin-postgresql

DISTFILES = \
	go.mod \
	go.sum \
	LICENSE \
	main.go \
	Makefile \
	postgresql.conf \
	README.md

DIST_SUBDIRS = \
	plugin \
	vendor

build:
	go build -o "$(PACKAGE)"

clean:
	rm -rf ./vendor
	rm -rf ./$(PACKAGE)*
	go clean ./...

check:
	go test -v -tags postgresql_tests ./...

style:
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) ./...

format:
	go fmt ./...

dist:
	go mod vendor; \
	major_verison=$$(grep 'const PLUGIN_VERSION_MAJOR' ./main.go | awk '{ print $$4 }'); \
	minor_verison=$$(grep 'const PLUGIN_VERSION_MINOR' ./main.go | awk '{ print $$4 }'); \
	patch_verison=$$(grep 'const PLUGIN_VERSION_PATCH' ./main.go | awk '{ print $$4 }'); \
	alphatag=$$(grep 'const PLUGIN_VERSION_RC' ./main.go | awk '{ print $$4 }' | cut -d '"' -f 2); \
	distdir="$(PACKAGE)-$${major_verison}.$${minor_verison}.$${patch_verison}$${alphatag}"; \
	dist_archive="$${distdir}.tar.gz"; \
	mkdir -p $${distdir}; \
	for distfile in '$(DISTFILES)'; do \
		cp -fp $${distfile} $${distdir}/; \
	done; \
	for subdir in '$(DIST_SUBDIRS)'; do \
		cp -fpR $${subdir} $${distdir}; \
	done; \
	tar -czvf $${dist_archive} $${distdir}; \
	rm -rf $${distdir}
