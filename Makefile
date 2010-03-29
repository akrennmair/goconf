include $(GOROOT)/src/Make.$(GOARCH)

TARG=conf
GOFILES=\
        conf.go\
	get.go\
	read.go\
	write.go

include $(GOROOT)/src/Make.pkg
