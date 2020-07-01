// +build !ignore_autogenerated

//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLogging) DeepCopyInto(out *CommonAuditLogging) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLogging.
func (in *CommonAuditLogging) DeepCopy() *CommonAuditLogging {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLogging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CommonAuditLogging) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingList) DeepCopyInto(out *CommonAuditLoggingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CommonAuditLogging, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingList.
func (in *CommonAuditLoggingList) DeepCopy() *CommonAuditLoggingList {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CommonAuditLoggingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingSpec) DeepCopyInto(out *CommonAuditLoggingSpec) {
	*out = *in
	in.Output.DeepCopyInto(&out.Output)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingSpec.
func (in *CommonAuditLoggingSpec) DeepCopy() *CommonAuditLoggingSpec {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingSpecHostAlias) DeepCopyInto(out *CommonAuditLoggingSpecHostAlias) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingSpecHostAlias.
func (in *CommonAuditLoggingSpecHostAlias) DeepCopy() *CommonAuditLoggingSpecHostAlias {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingSpecHostAlias)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingSpecOutput) DeepCopyInto(out *CommonAuditLoggingSpecOutput) {
	*out = *in
	out.Splunk = in.Splunk
	out.QRadar = in.QRadar
	if in.HostAliases != nil {
		in, out := &in.HostAliases, &out.HostAliases
		*out = make([]CommonAuditLoggingSpecHostAlias, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingSpecOutput.
func (in *CommonAuditLoggingSpecOutput) DeepCopy() *CommonAuditLoggingSpecOutput {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingSpecOutput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingSpecQRadar) DeepCopyInto(out *CommonAuditLoggingSpecQRadar) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingSpecQRadar.
func (in *CommonAuditLoggingSpecQRadar) DeepCopy() *CommonAuditLoggingSpecQRadar {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingSpecQRadar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingSpecSplunk) DeepCopyInto(out *CommonAuditLoggingSpecSplunk) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingSpecSplunk.
func (in *CommonAuditLoggingSpecSplunk) DeepCopy() *CommonAuditLoggingSpecSplunk {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingSpecSplunk)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditLoggingStatus) DeepCopyInto(out *CommonAuditLoggingStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditLoggingStatus.
func (in *CommonAuditLoggingStatus) DeepCopy() *CommonAuditLoggingStatus {
	if in == nil {
		return nil
	}
	out := new(CommonAuditLoggingStatus)
	in.DeepCopyInto(out)
	return out
}
