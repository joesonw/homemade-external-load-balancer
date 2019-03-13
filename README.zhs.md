
[English](./README.md) | 简体中文

## 介绍
![Diagram](/images/diagram.png)
![Diagram](/images/screenshot.png)

HELB (Homemade External Load-Balancer) 可以实现往常在云上才能看到的功能(AWS ELB + EKS or ALIYUN SLB + CSK)

HELB能做些什么?

* 根据k8s节点进行请求负载均衡 ([why?](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#caveats-and-limitations-when-preserving-source-ips))
* 用子域名来对服务进行DNS解析
* 给暴露出来的服务提供自定义域名（含解析）访问，避免了硬记一些ip和端口号

## 安装

> 请期待更多的例子
 目前请前往 [install](/install) 查看如何部署到k8s集群

Because kubernetes pods requires port usage declaration upon creation, you port usages in services (see below) are limited to what you specified when you started up `HELB`

If you want unlimited access, try clone the repo and run `go -o helb build cli/balance` to build your own copy for bare-metal deployment.

## Example

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    helb.v1alpha/alias: code
    helb.v1alpha/cert: default/codessl
    helb.v1alpha/protocol: http
    helb.v1alpha/secure-ports: https
  name: code-server
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 8443
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: code-server
  sessionAffinity: None
  type: LoadBalancer

```

