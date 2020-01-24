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

func (s *MySuite) TestKmakeRunJobIsKmakeRunOperation(c *C) {
	v := v1.KmakeRunJob{}
	var i interface{} = v
	_, ok := i.(KmakeRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeRunDummyIsKmakeRunOperation(c *C) {
	v := v1.KmakeRunDummy{}
	var i interface{} = v
	_, ok := i.(KmakeRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeRunFileWaitIsKmakeRunOperation(c *C) {
	v := v1.KmakeRunFileWait{}
	var i interface{} = v
	_, ok := i.(KmakeRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleRunStartIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleRunStart{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleRunStoptIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleRunStop{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleDeleteIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleDelete{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleCreateIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleCreate{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleResetIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleReset{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}

func (s *MySuite) TestKmakeScheduleForceIsKmakeScheduleRunOperation(c *C) {
	v := v1.KmakeScheduleForce{}
	var i interface{} = v
	_, ok := i.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, false)

	var p interface{} = &v
	_, ok = p.(KmakeScheduleRunOperation)
	c.Assert(ok, Equals, true)
}
