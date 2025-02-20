FROM golang:1.19-alpine AS builder
WORKDIR /backend
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/backend

FROM --platform=$BUILDPLATFORM node:18.12-alpine3.16 AS client-builder
WORKDIR /ui
COPY ui .
RUN npm install
RUN npm run build

FROM alpine:3.16
LABEL org.opencontainers.image.title="kubearchinspect"
COPY --from=builder /bin/backend /
COPY docker-compose.yaml .
COPY metadata.json .
COPY --from=client-builder /ui/build ui
COPY kubearchinspect.svg .

ENTRYPOINT ["/backend"]