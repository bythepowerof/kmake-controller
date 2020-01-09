# kmake-controller - WIP

A kubebuilder based controller to run `Makefiles` inside a cluster.

The idea is to shrink-wrap configuration into docker images, deploy them onto a cluster (optionally mutating the config) and then run `make` targets inside the cluster.

So multiple clusters can be supported from the same container with the resource mutated with `kustomize`


### Build status

| Platform    | CI Status | Coverage | Report Card | Documentation |
|------------|:-------|:------------|:------- |----------- |
linux       | [![Build Status](https://travis-ci.org/bythepowerof/kmake-controller.svg?branch=master)](https://travis-ci.org/bythepowerof/kmake-controller) | [![codecov](https://codecov.io/gh/bythepowerof/kmake-controller/branch/master/graph/badge.svg)](https://codecov.io/gh/bythepowerof/kmake-controller) | [![Go Report Card](https://goreportcard.com/badge/github.com/bythepowerof/kmake-controller)](https://goreportcard.com/report/github.com/bythepowerof/kmake-controller) | [![GoDoc](https://godoc.org/github.com/bythepowerof/kmake-controller?status.svg)](https://godoc.org/github.com/bythepowerof/kmake-controller) |







### Process

* Load the CRD defintion
* Run the controller on your cluster
* add [kmake.mk][2] to your `Makefile`
* Convert your `Makefile` into yaml with [pymake -y][1]
* Add the yaml into your `kmake` definition
* Load the definition into your cluster
* Create `kmake-run` definitions
* Run the definitions into your cluster

### How it works

`kmake` creates:-
* a persistent volume claim
* an env config map of any variables in the `Makefile`
* a config map containing the target part of the `Makefile` as yaml and a reduced `Makefile`

A first run of `kmake-run` will populate the PVC from the source docker image using the target defined in [kmake.mk][2]


### TODO

* Write `kmake-run` controller
* Improve the readme ;)
* Write some tests


[1]: https://github.com/bythepowerof/pymake
[2]: docs/kmake.mk
