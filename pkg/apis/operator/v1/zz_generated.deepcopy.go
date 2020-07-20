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
func (in *CommonAudit) DeepCopyInto(out *CommonAudit) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAudit.
func (in *CommonAudit) DeepCopy() *CommonAudit {
	if in == nil {
		return nil
	}
	out := new(CommonAudit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CommonAudit) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditList) DeepCopyInto(out *CommonAuditList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CommonAudit, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditList.
func (in *CommonAuditList) DeepCopy() *CommonAuditList {
	if in == nil {
		return nil
	}
	out := new(CommonAuditList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CommonAuditList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpec) DeepCopyInto(out *CommonAuditSpec) {
	*out = *in
	out.Fluentd = in.Fluentd
	in.Outputs.DeepCopyInto(&out.Outputs)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpec.
func (in *CommonAuditSpec) DeepCopy() *CommonAuditSpec {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecFluentd) DeepCopyInto(out *CommonAuditSpecFluentd) {
	*out = *in
	out.Resources = in.Resources
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecFluentd.
func (in *CommonAuditSpecFluentd) DeepCopy() *CommonAuditSpecFluentd {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecFluentd)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecHostAliases) DeepCopyInto(out *CommonAuditSpecHostAliases) {
	*out = *in
	if in.Hostnames != nil {
		in, out := &in.Hostnames, &out.Hostnames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecHostAliases.
func (in *CommonAuditSpecHostAliases) DeepCopy() *CommonAuditSpecHostAliases {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecHostAliases)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecOutputs) DeepCopyInto(out *CommonAuditSpecOutputs) {
	*out = *in
	out.Splunk = in.Splunk
	out.Syslog = in.Syslog
	if in.HostAliases != nil {
		in, out := &in.HostAliases, &out.HostAliases
		*out = make([]CommonAuditSpecHostAliases, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecOutputs.
func (in *CommonAuditSpecOutputs) DeepCopy() *CommonAuditSpecOutputs {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecOutputs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecRequirements) DeepCopyInto(out *CommonAuditSpecRequirements) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecRequirements.
func (in *CommonAuditSpecRequirements) DeepCopy() *CommonAuditSpecRequirements {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecResources) DeepCopyInto(out *CommonAuditSpecResources) {
	*out = *in
	out.Requests = in.Requests
	out.Limits = in.Limits
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecResources.
func (in *CommonAuditSpecResources) DeepCopy() *CommonAuditSpecResources {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecSplunk) DeepCopyInto(out *CommonAuditSpecSplunk) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecSplunk.
func (in *CommonAuditSpecSplunk) DeepCopy() *CommonAuditSpecSplunk {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecSplunk)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditSpecSyslog) DeepCopyInto(out *CommonAuditSpecSyslog) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditSpecSyslog.
func (in *CommonAuditSpecSyslog) DeepCopy() *CommonAuditSpecSyslog {
	if in == nil {
		return nil
	}
	out := new(CommonAuditSpecSyslog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonAuditStatus) DeepCopyInto(out *CommonAuditStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonAuditStatus.
func (in *CommonAuditStatus) DeepCopy() *CommonAuditStatus {
	if in == nil {
		return nil
	}
	out := new(CommonAuditStatus)
	in.DeepCopyInto(out)
	return out
}