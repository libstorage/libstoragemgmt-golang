#!/bin/bash

# # To debug, try:
# podman pull fedora
# podman run --privileged --rm=false --tty=true --interactive=true \
#    -v `pwd`:/libstoragemgmt-golang:rw fedora \
#    /bin/bash -c /libstoragemgmt-golang/test/docker_travis_test.sh
if [ "CHK$(rpm -E "%{?fedora}")" != "CHK" ];then
    dnf install python3-six golang libstoragemgmt libstoragemgmt-devel -y || exit 1
elif [ "CHK$(rpm -E "%{?el8}")" != "CHK" ];then
    dnf install python3-six golang libstoragemgmt libstoragemgmt-devel -y || exit 1
elif [ "CHK$(rpm -E "%{?el7}")" != "CHK" ];then
    # epel needed for golang
    yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y || exit 1
    yum install python-six golang libstoragemgmt libstoragemgmt-devel -y || exit 1
else
    echo "Unsupported distribution"
    exit 1
fi

mkdir -p /tmp/go/src/github.com/libstorage || exit 1
cp -av /libstoragemgmt-golang /tmp/go/src/github.com/libstorage/. || exit 1
cd /tmp/go/src/github.com/libstorage/libstoragemgmt-golang || exit 1

# Speed up tests
export LSM_SIM_TIME=0

# Start the service
/usr/bin/lsmd || exit 1

# Let the service get ready
sleep 5 || exit 1

# Make sure things are sane
lsmcli list --type pools -u simc:// || exit 1
lsmcli list --type plugins -u simc:// || exit 1

export GOPATH=/tmp/go

# Get the required lib for unit test
go get github.com/stretchr/testify/assert || exit 1

cd test || exit 1
./cov.sh || exit 1
