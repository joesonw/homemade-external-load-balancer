
English | [简体中文](./README.zhs.md)

## Introduction

At work, we all have AWS ELB, Aliyun SLB, etc. What if we just want to have a k8s cluster up running at home with this nice feature?

I give yout **HELB**(Homemade External Load-Balancer). It provides dynamic resolved dns names for your services on your machine, accessible private or public.


## How to use?

This is a simple demonstration of how it could be deployed (at least how I deployed).


My scenario was having a domain name, and I want to load balance my services/apps with dynamic resolved dns rather than configure it each time by hand.


Because I access it through LAN/OpenVPN only, I set it up with a private LAN ip (if you want to use it on the public internet, you could implement ip.Interface by requesting public ip services for the proxy to use)


I have a `traefik` reverse proxy for all the requests. Services will be found, resolved ad route to `traefik`.


In any case I have my ip changed, or your internet is ADSL (changes ip each time?). There is a nameserver module where it monitors ip, and update it to DNS providers periodically.


see [install](./install/static-dnspod-traefik) for a basic deployment


![Diagram](/images/diagram.jpg)


Following is how to annotate services so they could be picked up by HELB

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    local-load-balancer/alias: aria2
    local-load-balancer/enable: "true"
    local-load-balancer/port: http
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
