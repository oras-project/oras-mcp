FROM --platform=$BUILDPLATFORM docker.io/library/node:22-alpine AS builder
ADD . /app
WORKDIR /app
RUN npm install --ignore-scripts
RUN npm run build

# remove unnecessary files in order to reduce the number of layers
FROM --platform=$BUILDPLATFORM docker.io/library/node:22-alpine AS packer
COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/package*.json /app/

FROM --platform=$BUILDPLATFORM ghcr.io/oras-project/oras:v1.3.0-beta.2 AS oras

FROM --platform=$BUILDPLATFORM docker.io/library/node:22-alpine
RUN apk --update add ca-certificates
COPY --from=oras /bin/oras /bin/oras
COPY --from=packer /app /oras-mcp-server
WORKDIR /oras-mcp-server
RUN npm install --production --ignore-scripts

ENTRYPOINT ["node", "dist/index.js"]
