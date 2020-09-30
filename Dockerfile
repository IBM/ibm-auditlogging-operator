
# Build the manager binary
FROM golang:1.14.7 as builder
ARG GOARCH

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

# Build
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.2-349

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
COPY --from=builder /workspace/manager .

# copy licenses
RUN mkdir /licenses
COPY LICENSE /licenses

# USER nonroot:nonroot
USER 1001

ENTRYPOINT ["/manager"]
