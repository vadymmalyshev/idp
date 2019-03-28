FROM golang as idp-build-deps
RUN mkdir /idp
WORKDIR /idp
COPY . .
RUN go mod vendor && \
    go mod tidy && \
    go build . && \
#    mv config/config.dev.yaml config/config.yaml && \
    echo '127.0.0.1 id.hiveon.local' >> /etc/hosts && \
    ls

FROM golang
RUN mkdir -p /idp
WORKDIR /idp
COPY --from=idp-build-deps /idp/idp .
#COPY --from=idp-build-deps /idp/config config/.
EXPOSE 3000
#ENTRYPOINT ["./idp"]
CMD ["./idp","serve"]
