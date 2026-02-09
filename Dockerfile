FROM golang:1.25.7-alpine3.23 AS build

WORKDIR /work

COPY go.mod* go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /usr/local/bin/manual-approval main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=build /usr/local/bin/manual-approval /usr/local/bin/manual-approval

ENTRYPOINT ["/usr/local/bin/manual-approval"]
