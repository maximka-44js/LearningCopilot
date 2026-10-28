FROM busybox:1.37.0-musl AS healthcheck

FROM quay.io/keycloak/keycloak:25.0.6

COPY --from=healthcheck /bin/busybox /usr/local/bin/busybox
