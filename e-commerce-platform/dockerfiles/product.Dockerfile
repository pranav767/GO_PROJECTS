FROM golang:1.24 AS build
WORKDIR /src/services/product-service
COPY . /src
RUN apt-get update && apt-get install -y git build-essential && \
	CGO_ENABLED=0 go build -o /product-service .

FROM gcr.io/distroless/base-debian11
COPY --from=build /product-service /product-service
EXPOSE 8081
ENTRYPOINT ["/product-service"]
