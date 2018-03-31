# Build the service
FROM golang:1.10-alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
# Pre-accept the disclaimer as a Docker user would also have to export the ports
RUN touch iamcrazy

CMD ["app"]
