FROM golang:1.14-buster as stage1
WORKDIR src/github.com/sjpotter/udf-fs
COPY / ./
RUN dpkg -i misc/*.deb
RUN GO111MODULE=on go build -a -tags netgo -ldflags '-w -extldflags "-static"' cmd/udf-fs.go

FROM scratch AS export-stage
COPY --from=stage1 /go/src/github.com/sjpotter/udf-fs/udf-fs .
