FROM golang:alpine AS builder
# Git is required to fetch the dependencies
RUN apk update && apk add --no-cache git
WORKDIR /build
COPY . .
RUN go get -d -v
RUN go build -o /build/datasubst


FROM scratch
COPY --from=builder /build/datasubst /datasubst
ENTRYPOINT ["/datasubst"]

