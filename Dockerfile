FROM scratch

COPY dumb-init /dumb-init
COPY platform-skvs /skvs

ENTRYPOINT ["/dumb-init", "/skvs", "--port", "80", "--webhook-url", "hook"]

EXPOSE 80
