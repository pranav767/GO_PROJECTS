FROM golang:1.24 AS build
WORKDIR /src/services/cart-service
COPY . /src
RUN apt-get update && apt-get install -y git build-essential && \
	CGO_ENABLED=0 go build -o /cart-service .

FROM gcr.io/distroless/base-debian11
COPY --from=build /cart-service /cart-service
EXPOSE 8083
ENTRYPOINT ["/cart-service"]
