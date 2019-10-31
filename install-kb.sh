#!/bin/bash
set -x
set -e

if [[ ! -d /usr/local/kubebuilder ]];then
	os=$(go env GOOS)
	arch=$(go env GOARCH)
	ver=2.0.0

	# download kubebuilder and extract it to tmp
	curl -sL https://go.kubebuilder.io/dl/${ver}/${os}/${arch} | tar -xz -C /tmp/

	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	sudo mv /tmp/kubebuilder_${ver}_${os}_${arch} /usr/local/kubebuilder
fi
