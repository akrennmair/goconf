include $(GOROOT)/src/Make.$(GOARCH)

TARG=goconf.googlecode.com/hg/
GOFILES=\
        conf.go\
	get.go\
	read.go\
	write.go

include $(GOROOT)/src/Make.pkg
