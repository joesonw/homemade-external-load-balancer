package app

import (
	"encoding/hex"
	"fmt"
	"log"
	"pkg/annotations"
	"strconv"

	"time"

	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (c *Client) startWatchCluster() {
	resync := time.Second * 600
	informerFactory := informers.NewSharedInformerFactory(c.client, resync)
	serviceInformer := informerFactory.Core().V1().Services()
	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				svc := obj.(*corev1.Service)
				c.handleService(svc)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldSvc := oldObj.(*corev1.Service)
				newSvc := newObj.(*corev1.Service)
				for _, ingress := range oldSvc.Status.LoadBalancer.Ingress {
					delete(c.serviceMapping, ingress.Hostname)
				}
				c.handleService(newSvc)
			},
			DeleteFunc: func(obj interface{}) {
				svc := obj.(*corev1.Service)
				for _, ingress := range svc.Status.LoadBalancer.Ingress {
					delete(c.serviceMapping, ingress.Hostname)
				}
			},
		},
		resync,
	)

	informerFactory.Start(wait.NeverStop)
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

	port := &svc.Spec.Ports[0]
	if alias != "" {
		p, err := strconv.ParseInt(alias, 10, 64)
		if err == nil {
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(p) {
					port = sp.DeepCopy()
				}
			}
		} else {
			for _, sp := range svc.Spec.Ports {
				if sp.Name == alias {
					port = sp.DeepCopy()
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

	_, err := c.client.CoreV1().Services(svc.Namespace).UpdateStatus(svc)
	if err != nil {
		log.Printf("unable to update service %s: %s", svc.Name, err.Error())
	}
	log.Printf("registered %s for service %s\n", aliasHostmame, svc.Name)
	log.Printf("registered %s for service %s\n", defaultHostname, svc.Name)
	c.serviceMapping[alias] = dm
	c.serviceMapping[hostname] = dm
	c.refreshProxy()
}
