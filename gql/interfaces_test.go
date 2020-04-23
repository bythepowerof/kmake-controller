package gql

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bythepowerof/kmake-controller/api/v1"
)

var _ = Describe("Interfaces", func() {
	Context("KmakeNowScheduler Is KmakeScheduler", func() {
		It("Should create successfully", func() {
			v := v1.KmakeNowScheduler{}
			var i interface{} = v
			_, ok := i.(KmakeScheduler)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeScheduler)
			Expect(ok).To(Equal(true))
		})
	})

	Context("KmakeNowScheduler Is KmakeObject", func() {
		It("Should create successfully", func() {
			v := v1.KmakeNowScheduler{}
			var i interface{} = v
			_, ok := i.(KmakeObject)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeObject)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeRun Is KmakeObject", func() {
		It("Should create successfully", func() {
			v := v1.KmakeRun{}
			var i interface{} = v
			_, ok := i.(KmakeObject)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeObject)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleRun Is KmakeObject", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleRun{}
			var i interface{} = v
			_, ok := i.(KmakeObject)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeObject)
			Expect(ok).To(Equal(true))
		})
	})
	Context("Kmake Is KmakeObject", func() {
		It("Should create successfully", func() {
			v := v1.Kmake{}
			var i interface{} = v
			_, ok := i.(KmakeObject)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeObject)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeRunJob Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeRunJob{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeRunDummy Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeRunDummy{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeRunFileWait Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeRunFileWait{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleRunStart Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleRunStart{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleRunStop Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleRunStop{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleDelete Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleDelete{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleCreate Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleCreate{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleReset Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleReset{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
	Context("KmakeScheduleForce Is KmakeRunOperation", func() {
		It("Should create successfully", func() {
			v := v1.KmakeScheduleForce{}
			var i interface{} = v
			_, ok := i.(KmakeRunOperation)
			Expect(ok).To(Equal(false))
		
			var p interface{} = &v
			_, ok = p.(KmakeRunOperation)
			Expect(ok).To(Equal(true))
		})
	})
})
