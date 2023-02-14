from golang:alpine

COPY . /app
RUN mkdir config
WORKDIR /app 
RUN go build -o /humanz-cni .
RUN chmod 755 /humanz-cni
WORKDIR cni-init
RUN go build -o /humanz-cni-init .
ENTRYPOINT /humanz-cni-init


