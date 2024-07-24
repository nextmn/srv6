prefix = /usr/local
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin
BASHCOMPLETIONSDIR = $(exec_prefix)/share/bash-completion/completions
IPROUTE2_RTPROTOS_D = /etc/iproute2/rt_protos.d
IPROUTE2_RTTABLES_D = /etc/iproute2/rt_tables.d


RM = rm -f
INSTALL = install -D
MAKE = make --no-print-directory

.PHONY: install uninstall update build clean default
default: build
build:
	go build
clean:
	go clean
reinstall: uninstall install
update:
	go mod tidy
install:
	$(INSTALL) srv6 $(DESTDIR)$(bindir)/srv6
	$(INSTALL) etc/iproute2/rt_protos.d/nextmn.conf $(DESTDIR)$(IPROUTE2_RTPROTOS_D)/nextmn.conf
	$(INSTALL) etc/iproute2/rt_tables.d/nextmn.conf $(DESTDIR)$(IPROUTE2_RTTABLES_D)/nextmn.conf
	$(INSTALL) bash-completion/completions/srv6 $(DESTDIR)$(BASHCOMPLETIONSDIR)/srv6
	@echo "================================="
	@echo ">> Now run the following command:"
	@echo "\tsource $(DESTDIR)$(BASHCOMPLETIONSDIR)/srv6"
	@echo "================================="
uninstall:
	$(RM) $(DESTDIR)$(bindir)/srv6
	$(RM) $(DESTDIR)$(BASHCOMPLETIONSDIR)/srv6
	$(RM) $(DESTDIR)$(IPROUTE2_RTPROTOS_D)/nextmn.conf
	$(RM) $(DESTDIR)$(IPROUTE2_RTTABLES_D)/nextmn.conf

dev-install:
	python3 -m venv env
	env/bin/pip install sqlfluff
lint:
	env/bin/sqlfluff lint
	@echo Checking generated files
	@go generate ./... && git status --porcelain=v2 | { ! { grep _gen.go > /dev/null && echo "Generated files were not up to date."; } } && echo "Generated files are up to date"

test-postgres:
	@docker pull postgres # avoid doing this under timeout
	@chmod o+r ./internal/database/database.sql # mandatory to allow postgres’s docker container to process the file
	@timeout 15  docker run --name test-nextmn-srv6 -e POSTGRES_PASSWORD=postgres -v ./internal/database/database.sql:/docker-entrypoint-initdb.d/database.sql:ro postgres 2>/dev/stdout| grep -m 1 'PostgreSQL init process complete' || { echo "Could not initialize postgres" ; $(MAKE) stop-postgres ; exit 1 ; } && $(MAKE) stop-postgres
stop-postgres:
	@echo "Shutting down postgres"
	@docker stop test-nextmn-srv6 >/dev/null
	@docker rm --force=true --volumes=true test-nextmn-srv6 >/dev/null
