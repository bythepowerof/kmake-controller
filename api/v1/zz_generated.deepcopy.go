// +build !ignore_autogenerated

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kmake) DeepCopyInto(out *Kmake) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kmake.
func (in *Kmake) DeepCopy() *Kmake {
	if in == nil {
		return nil
	}
	out := new(Kmake)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Kmake) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeList) DeepCopyInto(out *KmakeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Kmake, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeList.
func (in *KmakeList) DeepCopy() *KmakeList {
	if in == nil {
		return nil
	}
	out := new(KmakeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KmakeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeRule) DeepCopyInto(out *KmakeRule) {
	*out = *in
	if in.Targets != nil {
		in, out := &in.Targets, &out.Targets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Commands != nil {
		in, out := &in.Commands, &out.Commands
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Prereqs != nil {
		in, out := &in.Prereqs, &out.Prereqs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeRule.
func (in *KmakeRule) DeepCopy() *KmakeRule {
	if in == nil {
		return nil
	}
	out := new(KmakeRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeRun) DeepCopyInto(out *KmakeRun) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeRun.
func (in *KmakeRun) DeepCopy() *KmakeRun {
	if in == nil {
		return nil
	}
	out := new(KmakeRun)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KmakeRun) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeRunList) DeepCopyInto(out *KmakeRunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KmakeRun, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeRunList.
func (in *KmakeRunList) DeepCopy() *KmakeRunList {
	if in == nil {
		return nil
	}
	out := new(KmakeRunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KmakeRunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeRunSpec) DeepCopyInto(out *KmakeRunSpec) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Targets != nil {
		in, out := &in.Targets, &out.Targets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeRunSpec.
func (in *KmakeRunSpec) DeepCopy() *KmakeRunSpec {
	if in == nil {
		return nil
	}
	out := new(KmakeRunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeRunStatus) DeepCopyInto(out *KmakeRunStatus) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeRunStatus.
func (in *KmakeRunStatus) DeepCopy() *KmakeRunStatus {
	if in == nil {
		return nil
	}
	out := new(KmakeRunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeSpec) DeepCopyInto(out *KmakeSpec) {
	*out = *in
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]KmakeRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Folders != nil {
		in, out := &in.Folders, &out.Folders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.JobTemplate.DeepCopyInto(&out.JobTemplate)
	in.PersistentVolumeClaimTemplate.DeepCopyInto(&out.PersistentVolumeClaimTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeSpec.
func (in *KmakeSpec) DeepCopy() *KmakeSpec {
	if in == nil {
		return nil
	}
	out := new(KmakeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KmakeStatus) DeepCopyInto(out *KmakeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KmakeStatus.
func (in *KmakeStatus) DeepCopy() *KmakeStatus {
	if in == nil {
		return nil
	}
	out := new(KmakeStatus)
	in.DeepCopyInto(out)
	return out
}
