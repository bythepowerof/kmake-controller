package gql

import (
	"github.com/bythepowerof/kmake-controller/api/v1"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestKmakeNowScheduler(c *C) {
	v := v1.KmakeNowScheduler{}
	var i interface{} = v
	_, ok := i.(KmakeScheduler)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduler)
	c.Assert(ok, Equals, true)
}
