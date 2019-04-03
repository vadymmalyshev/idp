FROM golang as idp-build-deps
RUN mkdir /idp
WORKDIR /idp
COPY . .
RUN go mod vendor && \
    go mod tidy && \
    go build . && \
    ls

FROM golang
RUN mkdir -p /idp
WORKDIR /idp
COPY --from=idp-build-deps /idp/idp .
#COPY --from=idp-build-deps /idp/config config/.
RUN     echo '127.0.0.1 id.hiveon.local hiveon.local' >> /etc/hosts
EXPOSE 3000
#ENTRYPOINT ["./idp"]
CMD ["./idp","-api=true"]
