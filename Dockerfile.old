FROM golang:1.11.5-alpine3.8

ARG NODE_ENV=production
ARG GIN_MODE=release
ARG ENV=production

ENV PORT=8081
ENV NODE_ENV=$NODE_ENV
ENV GIN_MODE=$GIN_MODE
ENV ENV=$ENV

RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
      git \
      build-base \
      python

ENV PATH /usr/local/bin:$PATH

RUN mkdir -p /opt
RUN mkdir -p /opt/idp

COPY . /opt/idp

WORKDIR /opt/idp

EXPOSE 3000

ENTRYPOINT [ "go", "run", "main.go" ]
CMD [ "serve" ]