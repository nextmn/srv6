prefix = /usr/local
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin
BASHCOMPLETIONSDIR = $(exec_prefix)/share/bash-completion/completions
IPROUTE2_RTPROTOS_D = /etc/iproute2/rt_protos.d
IPROUTE2_RTTABLES_D = /etc/iproute2/rt_tables.d


RM = rm -f
INSTALL = install -D

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
lint:
	sqlfluff lint
