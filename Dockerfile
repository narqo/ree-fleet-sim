FROM golang:1.15-alpine as builder
RUN apk add --update --no-cache build-base ca-certificates
WORKDIR /go/src/ree-fleet-sim
COPY . /go/src/ree-fleet-sim
RUN CGO_ENABLED=0 go build -a -trimpath -ldflags '-s -w -extldflags "-static"' ./cmd/fleetstate-server
RUN CGO_ENABLED=0 go build -a -trimpath -ldflags '-s -w -extldflags "-static"' ./cmd/simulator

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/ree-fleet-sim/fleetstate-server /bin/
COPY --from=builder /go/src/ree-fleet-sim/simulator /bin/
CMD ["/bin/fleetstate-server"]
