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

func (s *MySuite) TestKmakeNowSchedulerIsKmakeScheduler(c *C) {
	v := v1.KmakeNowScheduler{}
	var i interface{} = v
	_, ok := i.(KmakeScheduler)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduler)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeNowSchedulerIsKmakeObject(c *C) {
	v := v1.KmakeNowScheduler{}
	var i interface{} = v
	_, ok := i.(KmakeObject)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeObject)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeIsKmakeObject(c *C) {
	v := v1.Kmake{}
	var i interface{} = v
	_, ok := i.(KmakeObject)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeObject)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeRunIsKmakeObject(c *C) {
	v := v1.KmakeRun{}
	var i interface{} = v
	_, ok := i.(KmakeObject)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeObject)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleRunIsKmakeObject(c *C) {
	v := v1.KmakeScheduleRun{}
	var i interface{} = v
	_, ok := i.(KmakeObject)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeObject)
	c.Assert(ok, Equals, true)
}
