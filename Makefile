PREFIX=/usr/local
DESTDIR=
GOFLAGS=
BINDIR=${PREFIX}/gbusd

BLDDIR = build
EXT=
ifeq (${GOOS},windows)
    EXT=.exe
endif

APPS = gbusd_serv 
all: $(APPS)


$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${GOFLAGS} -o $@ ./example/$*

$(APPS): %: $(BLDDIR)/%
	@mkdir -p $(BLDDIR)/logs
	@cp -r example/gbusd_serv/conf          $(BLDDIR)/
	@cp    example/gbusd_serv/*.sh          $(BLDDIR)/

clean:
	rm -fr $(BLDDIR)/logs
	rm -fr $(BLDDIR)/conf
	rm -fr $(BLDDIR)

.PHONY: install clean all
.PHONY: $(APPS)

install: $(APPS)
	install -m 755 -d ${DESTDIR}${BINDIR}
	for APP in $^ ; do install -m 755 ${BLDDIR}/$$APP ${DESTDIR}${BINDIR}/$$APP${EXT} ; done
	@mkdir -p ${DESTDIR}${BINDIR}/logs

	@cp -r example/gbusd_serv/conf          ${DESTDIR}${BINDIR}/
	@cp    example/gbusd_serv/*.sh          ${DESTDIR}${BINDIR}/

	@echo  ${DESTDIR}${BINDIR}/
	@ls -al    ${DESTDIR}${BINDIR}/
