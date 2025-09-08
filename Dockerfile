# Copyright The ORAS Authors.
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

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.1-alpine AS builder
ARG TARGETPLATFORM
RUN apk add git make
ENV CLI_PKG=/oras-mcp
ADD . ${CLI_PKG}
WORKDIR ${CLI_PKG}
RUN make "build-$(echo $TARGETPLATFORM | tr / -)"
RUN mv ${CLI_PKG}/bin/${TARGETPLATFORM}/oras-mcp /go/bin/oras-mcp

FROM --platform=$BUILDPLATFORM ghcr.io/oras-project/oras:v1.3.0 AS oras

FROM docker.io/library/alpine:3.22.1
RUN apk --update add ca-certificates
COPY --from=oras /bin/oras /bin/oras
COPY --from=builder /go/bin/oras-mcp /bin/oras-mcp
RUN mkdir /workspace
WORKDIR /workspace
ENTRYPOINT  ["/bin/oras-mcp"]
