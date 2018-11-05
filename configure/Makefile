pbSource := $(shell echo *.proto)
#%.pb.go:%.proto
#	protoc --go_out=. $^
installpb:
			mkdir ~/pbinst; cd ~/pbinst; curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip; unzip protoc-3.6.1-linux-x86_64.zip -d protoc3; mv protoc3/bin/* $(GOPATH)/bin/
pb: clean installpb
	protoc --go_out=. $(pbSource)
	@echo "Done"

clean:
	-rm *.pb.go
