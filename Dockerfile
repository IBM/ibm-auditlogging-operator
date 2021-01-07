#
# Copyright 2021 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

FROM golang:1.14.7 as builder
ARG GOARCH=amd64

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY version/ version/

# Build
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM registry.access.redhat.com/ubi8/ubi-minimal@sha256:4a02f98dd98f11f0ab2849a57dcf4d9e33e24568e9e4ce81b549b60191b6c7d7

ARG VCS_REF
ARG VCS_URL

LABEL org.label-schema.vendor="IBM" \
  org.label-schema.name="ibm-auditlogging-operator" \
  org.label-schema.description="IBM Cloud Platform Common Services Audit Logging Component" \
  org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url=$VCS_URL \
  org.label-schema.license="Licensed Materials - Property of IBM" \
  org.label-schema.schema-version="1.0" \
  name="ibm-auditlogging-operator" \
  vendor="IBM" \
  description="IBM Cloud Platform Common Services Audit Logging Component" \
  summary="Audit Logging Service that forwards a service's audit logs to a SIEM."

WORKDIR /
# install operator binary
COPY --from=builder /workspace/manager .

# copy licenses
RUN mkdir /licenses
COPY LICENSE /licenses

# USER nonroot:nonroot
USER 1001

ENTRYPOINT ["/manager"]
