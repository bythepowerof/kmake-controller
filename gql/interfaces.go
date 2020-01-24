// +kubebuilder:object:generate=false
package gql

import (
	"github.com/bythepowerof/kmake-controller/api/v1"
)

type KmakeObject interface {
	GetName() string
	GetNamespace() string
}
type KmakeScheduler interface {
	GetName() string
	GetNamespace() string
	Variables() []v1.KV
	Monitor() []string
}

type KmakeRunOperation interface {
	Dummy() string
}

type KmakeScheduleRunOperation interface {
	Dummy() string
}
