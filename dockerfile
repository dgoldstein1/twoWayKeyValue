FROM golang:1.9

# setup go
ENV GOBIN $GOPATH/bin
ENV PATH $GOBIN:/usr/local/go/bin:$PATH

COPY build $GOBIN

# set docs
RUN mkdir /docs
COPY api/index.html /docs/index.html
ENV GRAPH_DOCS_DIR="/docs/*"
# enable zip
RUN apt-get update
RUN apt-get install -y zip
RUN zip -v
RUN mkdir -p /db/twowaykv
ENV GRAPH_DB_STORE_DIR="/db/twowaykv"
# set env
ENV GRAPH_DB_STORE_PORT="5001"

ENV COMMAND "serve"
RUN twowaykv --version
CMD twowaykv $COMMAND
