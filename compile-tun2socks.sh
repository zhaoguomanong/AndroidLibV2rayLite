#!/bin/bash

# Set magic variables for current file & dir
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__file="${__dir}/$(basename "${BASH_SOURCE[0]}")"
__base="$(basename ${__file} .sh)"

TMPDIR=$(mktemp -d)
cd $TMPDIR
cp $__dir/tun2socks.mk .
git clone --depth=1 https://github.com/shadowsocks/badvpn.git
git clone --depth=1 https://github.com/shadowsocks/libancillary.git

$NDK_HOME/ndk-build \
	NDK_PROJECT_PATH=. \
	APP_BUILD_SCRIPT=./tun2socks.mk \
	APP_ABI=all \
	APP_PLATFORM=android-19 \
	NDK_LIBS_OUT=$TMPDIR/libs \
	NDK_OUT=$TMPDIR/tmp \
	APP_SHORT_COMMANDS=false LOCAL_SHORT_COMMANDS=false -B -j4

install -v -m755 libs/armeabi-v7a/tun2socks  $__dir/shippedBinarys/ArchDep/arm/ 
install -v -m755 libs/arm64-v8a/tun2socks    $__dir/shippedBinarys/ArchDep/arm64/
install -v -m755 libs/x86/tun2socks          $__dir/shippedBinarys/ArchDep/386/ 
install -v -m755 libs/x86_64/tun2socks       $__dir/shippedBinarys/ArchDep/amd64/ 

cd $__dir/shippedBinarys
make clean && make shippedBinary

rm -rf $TMPDIR