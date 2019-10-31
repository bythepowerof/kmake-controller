#!/bin/bash
set -x
set -e

if [[ ! -d /usr/local/kubebuilder ]];then
	os=$(go env GOOS)
	arch=$(go env GOARCH)

	# download kubebuilder and extract it to tmp
	curl -sL https://go.kubebuilder.io/dl/2.0.0/${os}/${arch} | tar -xz -C /tmp/

	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	sudo mv /tmp/kubebuilder_2.0.1_${os}_${arch} /usr/local/kubebuilder
fi
