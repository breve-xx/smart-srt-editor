FROM golang:1.24.1 AS builder
RUN go install github.com/a-h/templ/cmd/templ@v0.3.833
ENV PATH="/go/bin:${PATH}"
ADD . /app
WORKDIR /app
RUN make && make generate && make build

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/bin/editor /editor
ENTRYPOINT ["/editor"]