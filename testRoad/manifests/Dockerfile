
# User golang based off al Alpine
FROM golang:latest

LABEL "vendor": "acme" \
      "microservice": "users" \
      "description": "Microservice for managing a backend persistent store for an object" \
      "version": "1.0.0" \
      "tempalte": "datamgr" \
      "definition": ".templates/microservices/ga/datamgr/definition.yaml" \
      "env": "dev"

MAINTAINER "support@pavedroad.io"

# Build paths for placing kevlar microservice
ENV ms users
ENV pavedroad /pavedroad
ENV pavedroadbin $pavedroad/$ms

# make working directory, move to it, and copy in prToken ms
RUN mkdir ${pavedroad}
WORKDIR ${pavedroad}
COPY $ms $pavedroad

EXPOSE 8081
CMD ["/bin/sh", "-c", "$pavedroadbin"]
