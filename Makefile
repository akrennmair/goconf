include $(GOROOT)/src/Make.inc

TARG=conf
GOFILES=\
        conf.go\
	get.go\
	read.go\
	write.go

include $(GOROOT)/src/Make.pkg
