FROM golang:latest
WORKDIR /golang-crawlstackoverflow-website
COPY . ./
RUN go mod tidy
RUN go build
CMD [ "./ormcrawldata" ]
