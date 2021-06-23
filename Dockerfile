# build stage
FROM debian:stretch-slim AS build-env
RUN apt-get update && apt install -y ca-certificates curl
RUN cd /usr/local && curl -LO https://golang.org/dl/go1.16.5.linux-amd64.tar.gz && \
  tar zxf go1.16.5.linux-amd64.tar.gz && rm -f go1.16.5.linux-amd64.tar.gz
RUN apt install -y make
ENV GOROOT=/usr/local/go
ENV PATH="${GOROOT}/bin:/root/go/bin:${PATH}"
ADD . /src
RUN cd /src && make todosvc

# final stage
FROM debian:stretch-slim
WORKDIR /app
COPY --from=build-env /src/dist/todosvc /usr/bin
CMD /usr/bin/todosvc
