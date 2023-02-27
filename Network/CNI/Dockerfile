from golang:alpine
RUN apk update && apk add --no-cache iptables 
COPY . /app
RUN mkdir config
WORKDIR /app/cmd/humanz-cni
RUN go build -o /humanz-cni .
RUN chmod 755 /humanz-cni
WORKDIR /app/cmd/humanz-cni-agent
RUN go build -o /humanz-cni-agent .
ENTRYPOINT /humanz-cni-agent


