FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
RUN mkdir -p /go/src/gitee.com/openeuler/osi-task-manager
COPY . /go/src/gitee.com/openeuler/osi-task-manager
#RUN cd /go/src/gitee.com/openeuler/osi-task-manager && CGO_ENABLED=1 go build -v -o ./osi-task-manager main.go

# copy binary config and utils
FROM golang:latest
RUN mkdir -p /opt/app/ && mkdir -p /opt/app/conf/
COPY ./conf/product_app.conf /opt/app/conf/app.conf
# overwrite config yaml
COPY  --from=BUILDER /go/src/gitee.com/openeuler/osi-task-manager/osi-task-manager /opt/app
RUN cd /opt/app && go build -v -o ./osi-task-manager main.go
COPY /go/src/gitee.com/openeuler/osi-task-manager/osi-task-manager/conf/product_app.conf /opt/app/conf/app.conf
WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/osi-task-manager"]