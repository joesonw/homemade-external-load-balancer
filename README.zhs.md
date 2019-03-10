
[English](./README.md) | 简体中文

## Introduction

工作当中，我们可以用到AWS ELB，阿里云SLB来访问我们的k8s集群服务。 如果我们自己搭建的话，又该如何呢？

通过使用**HELB**(Homemade External Load-Balancer)，可以让你在自己的机器/集群上取得同样的效果

## How to use?

下面是一个简单的实例（我如何部署的）

我自己有一个域名。在这个域名的基础上，我解析了一个二级域名的NS记录到集群，然后通过集群来做动态的三级域名解析

最后的结果是把可用的服务解析到Traefik的服务上， 通过Traefik来代理服务

![Diagram](/images/diagram.jpg)


下面是一个如何让给服务被HELB识别出来的例子：

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    helb/alias: aria2
    helb/enable: "true"
    helb/port: http
  name: aria2
  namespace: default
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 6800
  selector:
    app: aria2
  type: LoadBalancer
```
