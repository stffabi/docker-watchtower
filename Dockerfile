FROM centurylink/ca-certs
MAINTAINER Fabrizio Steiner <stffabi@users.noreply.github.com>
LABEL "com.centurylinklabs.watchtower"="true"

COPY watchtower /
ENTRYPOINT ["/watchtower"]
