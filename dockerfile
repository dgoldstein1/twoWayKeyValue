from python:2.7

RUN pip install awscli

COPY build/* /bin

# set docs
RUN mkdir /docs
COPY api/index.html /docs/index.html
COPY VERSION /docs/VERSION

# set env
ENV GRAPH_DOCS_DIR="/docs/*"
RUN mkdir -p /db/twowaykv
ENV GRAPH_DB_STORE_DIR="/db/twowaykv"
ENV GRAPH_DB_STORE_PORT="5001"
ENV COMMAND "serve"

COPY docker_run.sh docker_run.sh
CMD ./docker_run.sh


