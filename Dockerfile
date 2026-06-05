# Start from the latest golang base image
FROM golang:1.25-alpine3.23 AS builder

ENV PATH /usr/local/go/bin:$PATH
ENV GOLANG_VERSION 1.25.5

# Add Maintainer Info
LABEL maintainer="cgil"


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY cmd/template4YourProjectNameServer ./template4YourProjectNameServer
COPY pkg ./pkg
COPY gen ./gen

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o template4YourProjectNameServer ./template4YourProjectNameServer


######## Start a new stage  #######
# using from scratch for size and security reason
# Containers Are Not VMs! Which Base Container (Docker) Images Should We Use?
# https://blog.baeke.info/2021/03/28/distroless-or-scratch-for-go-apps/
# https://github.com/vfarcic/base-container-images-demo
# https://youtu.be/82ZCJw9poxM
FROM scratch
# to comply with security best practices
# Running containers with 'root' user can lead to a container escape situation (the default with Docker...).
# It is a best practice to run containers as non-root users
# https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
# https://docs.docker.com/engine/reference/builder/#user
USER 1221:1221
WORKDIR /goapp

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/template4YourProjectNameServer .

ENV PORT="${PORT}"
ENV DB_DRIVER="${DB_DRIVER}"
ENV DB_HOST="${DB_HOST}"
ENV DB_PORT="${DB_PORT}"
ENV DB_NAME="${DB_NAME}"
ENV DB_USER="${DB_USER}"
ENV DB_PASSWORD="${DB_PASSWORD}"
ENV DB_SSL_MODE="${DB_SSL_MODE}"
ENV JWT_SECRET="${JWT_SECRET}"
ENV JWT_ISSUER_ID="${JWT_ISSUER_ID}"
ENV JWT_CONTEXT_KEY="${JWT_CONTEXT_KEY}"
ENV JWT_DURATION_MINUTES="${JWT_DURATION_MINUTES}"
ENV ADMIN_USER="${ADMIN_USER}"
ENV ADMIN_EMAIL="${ADMIN_EMAIL}"
ENV ADMIN_ID="${ADMIN_ID}"
ENV ADMIN_EXTERNAL_ID="${ADMIN_EXTERNAL_ID}"
ENV ADMIN_PASSWORD="${ADMIN_PASSWORD}"
ENV ALLOWED_HOSTS="${ALLOWED_HOSTS}"
ENV APP_ENV="${APP_ENV}"
# Expose port  to the outside world, template4YourProjectName will use the env PORT as listening port or 8080 as default
EXPOSE 9090

# how to check if container is ok https://docs.docker.com/engine/reference/builder/#healthcheck
HEALTHCHECK --start-period=5s --interval=30s --timeout=3s \
    CMD curl --fail http://localhost:9090/health || exit 1


# Command to run the executable
CMD ["./template4YourProjectNameServer"]
