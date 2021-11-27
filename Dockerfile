FROM golang:latest as BUILDER
LABEL maintainer="zhangjianjun"

# build binary
RUN mkdir -p /go/src/gitee.com/openeuler/osi-task-manager
COPY . /go/src/gitee.com/openeuler/osi-task-manager
RUN cd /go/src/gitee.com/openeuler/osi-task-manager && CGO_ENABLED=1 go build -v -o ./osi-task-manager main.go

# copy binary config and utils
FROM openeuler/openeuler:21.03
RUN mkdir -p /opt/app/conf/
COPY ./conf/product_app.conf /opt/app/conf/app.conf
# overwrite config yaml
COPY --from=BUILDER /go/src/gitee.com/openeuler/osi-task-manager/osi-task-manager /opt/app
WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/osi-task-manager"]