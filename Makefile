PACKAGE=zabbix-agent2-plugin-postgresql

DISTFILES = \
	go.mod \
	go.sum \
	LICENSE \
	main.go \
	Makefile \
	postgres.conf \
	README.md

DIST_SUBDIRS = \
	plugin \
	vendor

build:
	go build -o "$(PACKAGE)"

clean:
	go clean ./...
	rm -rf ./vendor
	rm -rf ./$(PACKAGE)*

check:
	go test -v -tags postgres_tests ./...

style:
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) ./...

format:
	go fmt ./...

dist:
	go mod vendor; \
	major_verison=$$(grep 'MajorVersion = ' ./vendor/git.zabbix.com/ap/plugin-support/plugin/comms/version.go | awk '{ print $$3 }'); \
	minor_verison=$$(grep 'MinorVersion = ' ./vendor/git.zabbix.com/ap/plugin-support/plugin/comms/version.go | awk '{ print $$3 }'); \
	plugin_verison=$$(grep 'pluginVersion =' ./main.go | awk '{ print $$4 }'); \
	distdir="$(PACKAGE)-$${major_verison}.$${minor_verison}.$${plugin_verison}"; \
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
