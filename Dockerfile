FROM gcr.io/distroless/static:nonroot

COPY manual-approval /usr/local/bin/manual-approval

ENTRYPOINT ["/usr/local/bin/manual-approval"]
