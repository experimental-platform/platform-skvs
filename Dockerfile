FROM experimentalplatform/ubuntu:latest

COPY platform-skvs /skvs

CMD ["dumb-init", "/skvs", "--port", "80", "--webhook-url", "hook"]

EXPOSE 80
