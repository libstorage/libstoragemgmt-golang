#!/bin/bash

# If we have an argument, it's the directory of the libstoragemgmt source code, so use it
# instead of running against installed packages

# To run locally, do:
# podman pull fedora
# Run the following in the root of the golang source tree
# podman run --privileged --rm=false --tty=true --interactive=true -v \
#    `pwd`:/libstoragemgmt-golang:rw fedora \
#    /bin/bash -c "cd /libstoragemgmt-golang && pwd && test/docker_travis_test.sh"

if [ -z "$@" ]; then
    if [ "CHK$(rpm -E "%{?fedora}")" != "CHK" ];then
        dnf install python3-six golang libstoragemgmt libstoragemgmt-devel git-core -y || exit 1
    elif [ "CHK$(rpm -E "%{?el8}")" != "CHK" ];then
        dnf install dnf-plugins-core -y || exit 1
	dnf config-manager --set-enabled ol8_codeready_builder -y || exit 1
        dnf install python3-six golang libstoragemgmt libstoragemgmt-devel git-core -y || exit 1
    elif [ "CHK$(rpm -E "%{?el7}")" != "CHK" ];then
        # epel needed for golang
        yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y || exit 1
        yum install python-six golang libstoragemgmt libstoragemgmt-devel git-core -y || exit 1
    else
        echo "Unsupported distribution"
        exit 1
    fi

    # Start the service
    /usr/bin/lsmd || exit 1

    # Let the service get ready
    sleep 5 || exit 1

    # Make sure things are sane
    lsmcli list --type pools -u simc:// || exit 1
    lsmcli list --type plugins -u simc:// || exit 1

else

    pwd

    # If we are running against a source directory, we assume daemon already running for us.
    export LSMSRC="$1"

    if [ "CHK$(rpm -E "%{?fedora}")" != "CHK" ];then
        dnf install golang sed -y || exit 1
    elif [ "CHK$(rpm -E "%{?el8}")" != "CHK" ];then
        dnf install dnf-plugins-core -y || exit 1
        dnf config-manager --set-enabled ol8_codeready_builder -y || exit 1
        dnf install golang sed -y || exit 1
    elif [ "CHK$(rpm -E "%{?el7}")" != "CHK" ];then
        # epel needed for golang
        yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y || exit 1
        yum install golang sed -y || exit 1
    else
        echo "Unable to run golang test for this distro, skipping..."
        exit 0
    fi


    # Remove the pkg-config until we find a better way to handle this
    # The env. variable apparently doesn't override this behavior.
    sed '/pkg-config/d' -i ./localdisk/localdisk.go || exit 1

fi

TESTING_DIR=/tmp/go/src/github.com/libstorage/libstoragemgmt-golang/
mkdir -p $TESTING_DIR || exit 1

# Circle places you at root of checkout
cp -av . $TESTING_DIR || exit 1
cd $TESTING_DIR || exit 1

# Speed up tests
export LSM_SIM_TIME=0

export GOPATH=/tmp/go

# Get the required lib for unit test
go get github.com/stretchr/testify/assert || exit 1

cd test || exit 1
./cov.sh || exit 1
