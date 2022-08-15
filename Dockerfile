# Start by building the gflication.
FROM golang:1.17-bullseye as build

WORKDIR /go/src/spotctl
ADD . /go/src/spotctl

RUN go get -d -v ./...

RUN go build -o /go/bin/spotctl

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian11
COPY --from=build /go/bin/spotctl /
ENTRYPOINT ["/spotctl"]
