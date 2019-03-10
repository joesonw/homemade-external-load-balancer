package app

import (
	"encoding/hex"
	"fmt"
	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"log"
	"pkg/annotations"
	"strconv"
)

func (c *Client) startWatchCluster() {
	services, err := c.client.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, svc := range services.Items {
		c.handleService(&svc)
	}

	watcher, err := c.client.CoreV1().Services(metav1.NamespaceAll).Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for r := range watcher.ResultChan() {
		svc := r.Object.(*corev1.Service)
		if r.Type == watch.Deleted {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				delete(c.serviceMapping, ingress.Hostname)
			}
		} else if r.Type == watch.Modified {
			c.handleService(svc)
		} else if r.Type == watch.Added {
			c.handleService(svc)
		}
	}

}

func (c *Client) handleService(svc *corev1.Service) {
	if svc.Annotations[annotations.Enable] != "true" {
		return
	}
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return
	}
	alias := svc.Annotations[annotations.Alias]
	buf := make([]byte, 32)
	hex.Encode(buf, uuid.NewV4().Bytes())
	hostname := string(buf)

	var port *corev1.ServicePort
	port = &svc.Spec.Ports[0]
	if alias != "" {
		p, err := strconv.ParseInt(alias, 10, 64)
		if err == nil {
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(p) {
					port = &sp
				}
			}
		} else {
			for _, sp := range svc.Spec.Ports {
				if sp.Name == alias {
					port = &sp
				}
			}
		}
	}

	if port == nil {
		return
	}

	basename := fmt.Sprintf("%s.%s", c.config.SubDomain, c.config.Domain)

	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		name := ingress.Hostname[0 : len(ingress.Hostname)-len(basename)-1]
		if name != alias {
			hostname = name
		}
	}

	dm := &dnsMapping{
		DNSName: fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace),
		Port:    port.Port,
	}

	aliasHostmame := fmt.Sprintf("%s.%s", alias, basename)
	defaultHostname := fmt.Sprintf("%s.%s", hostname, basename)

	svc.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					Hostname: aliasHostmame,
				},
				{
					Hostname: defaultHostname,
				},
			},
		},
	}

	c.client.CoreV1().Services(svc.Namespace).UpdateStatus(svc)
	log.Printf("registered %s for service %s\n", aliasHostmame, svc.Name)
	log.Printf("registered %s for service %s\n", defaultHostname, svc.Name)
	c.serviceMapping[alias] = dm
	c.serviceMapping[hostname] = dm
	c.refreshProxy()
}
