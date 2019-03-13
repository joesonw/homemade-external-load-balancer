package app

import (
	"encoding/hex"
	"fmt"
	"log"
	"pkg/annotations"

	"time"

	"strings"

	uuid "github.com/satori/go.uuid"
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
				if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
					return
				}
				c.serviceCacheLock.RLock()
				version := c.serviceCache[svc.Name]
				c.serviceCacheLock.RUnlock()
				if version == svc.ResourceVersion {
					return
				}
				c.addService(svc)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldSvc := oldObj.(*corev1.Service)
				newSvc := newObj.(*corev1.Service)
				if oldSvc.Spec.Type != corev1.ServiceTypeLoadBalancer {
					return
				}
				c.serviceCacheLock.RLock()
				version := c.serviceCache[newSvc.Name]
				c.serviceCacheLock.RUnlock()
				if version == newSvc.ResourceVersion {
					return
				}
				c.removeService(oldSvc)
				c.addService(newSvc)
			},
			DeleteFunc: func(obj interface{}) {
				svc := obj.(*corev1.Service)
				if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
					return
				}
				c.removeService(svc)
			},
		},
		resync,
	)

	informerFactory.Start(wait.NeverStop)
}

func (c *Client) removeService(svc *corev1.Service) {
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		c.dns.Add(ingress.Hostname)
		c.dns.Add(ingress.Hostname)
		c.proxy.RemoveService(ingress.Hostname)
		log.Printf("removed %s service %s\n", ingress.Hostname, svc.Name)
	}
}

func (c *Client) addService(svc *corev1.Service) {
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return
	}

	protocol := svc.Annotations[annotations.Protocol]
	if protocol != "http" {
		return
	}

	alias := strings.TrimSpace(svc.Annotations[annotations.Alias])
	buf := make([]byte, 32)
	hex.Encode(buf, uuid.NewV4().Bytes())
	hostname := string(buf)

	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if !strings.HasPrefix(ingress.Hostname, alias) {
			hostname = strings.Split(ingress.Hostname, ".")[0]
		}
	}

	basename := fmt.Sprintf("%s.%s", c.config.SubDomain, c.config.Domain)

	aliasHostmame := fmt.Sprintf("%s.%s", alias, basename)
	defaultHostname := fmt.Sprintf("%s.%s", hostname, basename)

	if err := c.proxy.AddService(svc, []string{defaultHostname, aliasHostmame}); err != nil {
		log.Printf("unable to add service to proxy for service %s: %s", svc.Name, err.Error())
		return
	}

	ingresses := []corev1.LoadBalancerIngress{{Hostname: defaultHostname}}
	if alias != "" {
		ingresses = append(ingresses, corev1.LoadBalancerIngress{Hostname: aliasHostmame})
		c.dns.Add(aliasHostmame)
	}

	svc.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: ingresses,
		},
	}

	c.dns.Add(defaultHostname)

	c.serviceCacheLock.Lock()
	newSvc, err := c.client.CoreV1().Services(svc.Namespace).UpdateStatus(svc)
	if err == nil {
		c.serviceCache[newSvc.Name] = newSvc.ResourceVersion
	}
	c.serviceCacheLock.Unlock()
	if err != nil {
		log.Printf("unable to update service %s: %s", svc.Name, err.Error())
	}
	if alias != "" {
		log.Printf("added %s service %s\n", aliasHostmame, svc.Name)
	}
	log.Printf("added %s service %s\n", defaultHostname, svc.Name)
}
