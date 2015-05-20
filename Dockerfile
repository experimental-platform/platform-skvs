FROM dockerregistry.protorz.net/ubuntu:latest

COPY skvs /skvs

CMD ["/skvs", "--port", "80", "--webhook-url", "hook"]

EXPOSE 80
