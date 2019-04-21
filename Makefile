pb:
	  go get -u github.com/golang/protobuf/protoc-gen-go
		@echo "pb Start"
asset:
	@echo $(shell ./gen_assets.sh)

shippedBinary:
	cd shippedBinarys; $(MAKE) shippedBinary

fetchDep:
	-go get  github.com/2dust/AndroidLibV2rayLite
	go get github.com/2dust/AndroidLibV2rayLite

ANDROID_HOME=$(HOME)/android-sdk-linux
export ANDROID_HOME
PATH:=$(PATH):$(GOPATH)/bin
export PATH
downloadGoMobile:
	go get golang.org/x/mobile/cmd/...
	sudo apt-get install -qq libstdc++6:i386 lib32z1 expect
	cd ~ ;curl -L https://raw.githubusercontent.com/2dust/AndroidLibV2rayLite/master/ubuntu-cli-install-android-sdk.sh | sudo bash - > /dev/null
	ls ~
	ls ~/android-sdk-linux/
	gomobile init ;gomobile bind -v  -tags json github.com/2dust/AndroidLibV2rayLite

BuildMobile:
	@echo Stub

all: asset pb shippedBinary fetchDep
	@echo DONE
