FROM golang:alpine AS build
WORKDIR /go/src
COPY . .
ENV CGO_ENABLED=0
RUN go install cmd/client/quotaservice-cli.go
RUN go install cmd/server/quotaservice.go

FROM scratch
COPY --from=build /go/bin/quotaservice /quotaservice
COPY --from=build /go/bin/quotaservice-cli /quotaservice-cli
COPY --from=build /go/src/admin/public /admin/public
EXPOSE 10990
CMD ["/quotaservice"]
