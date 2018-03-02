FROM golang:1.9-alpine AS builder
WORKDIR /go/src/github.com/edgexfoundry/edgex-go
RUN apk update && apk add make
COPY . .
RUN make cmd/core-command/core-command

FROM scratch

ENV APP_PORT=48082
#expose command data port
EXPOSE $APP_PORT

WORKDIR /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/core-command/core-command /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/core-command/res/configuration-docker.json /res/configuration.json
ENTRYPOINT ["/core-command"]
