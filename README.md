
 English | [简体中文](./README.zhs.md)

## Introduction
![Diagram](/images/diagram.png)
![Diagram](/images/screenshot.png)

HELB (Homemade External Load-Balancer) does what you would normally see in cloud providers (AWS ELB + EKS or ALIYUN SLB + CSK)

## What does HELB do ?

* Load balance traffic across nodes ([why?](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#caveats-and-limitations-when-preserving-source-ips))
* Resolve sub-domain DNS for services
* Assign custom domain names to services, access services without memorizing ip and ports.

## Install

> Working on better examples
 see [install](/install) for deployment into kubernetes cluster

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

