#!/bin/bash
set -x
set -e

if [[ ! -d /usr/local/kubebuilder ]];then
	os=$(go env GOOS)
	arch=$(go env GOARCH)
	version=2.2.0

	if [[ ! -f /tmp/kubebuilder/kubebuilder_${version}_${os}_${arch} ]];then

		# created by travis cache after first run
		mkdir -p /tmp/kubebuilder

		# download kubebuilder and extract it to tmp
		curl -sL https://go.kubebuilder.io/dl/${version}/${os}/${arch} | tar -xz -C /tmp/kubebuilder
	fi

	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	sudo cp -r /tmp/kubebuilder/kubebuilder_${version}_${os}_${arch} /usr/local/kubebuilder
fi
