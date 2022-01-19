FROM gcr.io/distroless/static

COPY ninjascepserver-linux-amd64 /ninjascepserver

EXPOSE 8083

ENTRYPOINT ["/ninjascepserver"]
