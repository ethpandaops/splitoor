FROM gcr.io/distroless/cc-debian12:latest
COPY splitoor* /splitoor
ENTRYPOINT ["/splitoor"]
