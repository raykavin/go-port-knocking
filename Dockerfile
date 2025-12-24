FROM golang:1.24.5-alpine3.22 AS build

ENV APP_NAME=${PROJECT_NAME}

WORKDIR /build

# Install build dependencies
RUN apk update && \
    apk add --no-cache git gcc musl-dev tzdata

# Set timezone
RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone

COPY . .

# Build the Go application with CGO enabled
RUN CGO_ENABLED=1 go build ${GO_BUILD_FLAGS} -o ${APP_NAME} cmd/api/main.go

FROM alpine:3.22 AS runtime

ENV APP_NAME=${PROJECT_NAME}
ARG APP_USER_ID=1001
ARG APP_GROUP_ID=1001
ARG TZ=America/Belem

ENV APP_HOME=/${APP_NAME} \
    LOG_DIR=/var/log/${APP_NAME}

# Install runtime dependencies
RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl wget bash

# Set timezone
RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone

# Create non-root user and group
RUN addgroup -g ${APP_GROUP_ID} ${APP_NAME} && \
    adduser -D -u ${APP_USER_ID} -G ${APP_NAME} -s /bin/sh ${APP_NAME} && \
    mkdir -p ${APP_HOME} ${LOG_DIR} && \
    chown -R ${APP_NAME}:${APP_NAME} ${APP_HOME} ${LOG_DIR}

# Copy the compiled binary from build stage
COPY --from=build --chown=${APP_NAME}:${APP_NAME} /build/${APP_NAME} /usr/local/bin/${APP_NAME}

# Copy healthcheck script if exists
# COPY --chown=${APP_NAME}:${APP_NAME} ./scripts/sh/healthcheck.sh /usr/local/bin/

# Make binaries executable
RUN if [ -f /usr/local/bin/healthcheck.sh ]; then \
        chmod +x /usr/local/bin/healthcheck.sh; \
    fi && \
    chmod +x /usr/local/bin/${APP_NAME}

# Use non-root user
USER ${APP_NAME}

WORKDIR ${APP_HOME}

# Define volumes for data and logs
VOLUME ["${APP_HOME}", "${LOG_DIR}"]

# Healthcheck configuration
# HEALTHCHECK --interval=60s --timeout=10s --start-period=5s --retries=3 \
#     CMD if [ -f /usr/local/bin/healthcheck.sh ]; then \
#             /usr/local/bin/healthcheck.sh; \
#         else \
#             /usr/local/bin/${APP_NAME} health || exit 1; \
#         fi

# Expose application port
# EXPOSE 3000

# Entrypoint 
ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/$APP_NAME -config $APP_HOME/config.yml"]
