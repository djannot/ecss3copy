FROM google/golang
WORKDIR /go/src
RUN git clone https://github.com/djannot/ecss3copy.git
WORKDIR /go/src/ecss3copy
RUN go get "github.com/djannot/ecss3copy/s3"
RUN go get "github.com/jessevdk/go-flags"
RUN go get "github.com/mitchellh/goamz/aws"
RUN go build .
