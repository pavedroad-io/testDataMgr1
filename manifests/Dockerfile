
# User golang based off al Alpine
FROM golang:latest

LABEL "io.pavedroad.vendor": "acme" \
      "io.pavedroad.microservice": "users" \
      "io.pavedroad.description": "Microservice for managing a backend persistent store for an object" \
      "io.pavedroad.version": "1.0.0" \
      "io.pavedroad.tempalte": "datamgr" \
      "io.pavedroad.definition": "testDataMgr.yaml" \
      "env": "dev"

MAINTAINER "support@pavedroad.io"

# Build paths for placing kevlar microservice
ENV ms users
ENV pavedroad /pavedroad
ENV pavedroadbin $pavedroad/$ms

# make working directory, move to it, and copy in prToken ms
RUN mkdir -p ${pavedroad}/logs
WORKDIR ${pavedroad}
COPY $ms $pavedroad

EXPOSE 8081
CMD ["/bin/sh", "-c", "$pavedroadbin"]
