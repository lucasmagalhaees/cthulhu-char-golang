# Build a static binary, then ship it alone: the image is a few MB and the container wakes in milliseconds.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api .

FROM gcr.io/distroless/static-debian12
COPY --from=build /api /api
EXPOSE 80
ENTRYPOINT ["/api"]
