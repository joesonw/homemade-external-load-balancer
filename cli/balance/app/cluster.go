package app

import (
	"encoding/hex"
	"fmt"
	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"pkg/annotations"
	"strconv"
	"time"
	"log"
)

func (c *Client) startWatchCluster() {
	watchList := cache.NewListWatchFromClient(c.client.CoreV1().RESTClient(), "services", metav1.NamespaceAll, fields.Everything())
	store, controller := cache.NewInformer(
		watchList,
		&corev1.Service{},
		time.Second*30,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.handleService(obj.(*corev1.Service))
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				oldSvc := oldObj.(*corev1.Service)
				newSvc := oldObj.(*corev1.Service)
				for _, ingress := range oldSvc.Status.LoadBalancer.Ingress {
					delete(c.serviceMapping, ingress.Hostname)
				}
				c.handleService(newSvc)
			},
		},
	)

	for _, obj := range store.List() {
		c.handleService(obj.(*corev1.Service))
	}

	stop := make(chan struct{})
	controller.Run(stop)
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
		name := ingress.Hostname[0:len(ingress.Hostname)-len(basename) - 1]
		if name != alias {
			hostname = name
		}
	}

	dm := &dnsMapping{
		DNSName: fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace),
		Port:    port.Port,
	}
	svc.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					Hostname: fmt.Sprintf("%s.%s", alias, basename),
				},
				{
					Hostname: fmt.Sprintf("%s.%s", hostname, basename),
				},
			},
		},
	}
	log.Printf("registered %s for service %s\n", fmt.Sprintf("%s.%s", alias, basename), svc.Name)
	log.Printf("registered %s for service %s\n", fmt.Sprintf("%s.%s", hostname, basename), svc.Name)
	c.client.CoreV1().Services(svc.Namespace).UpdateStatus(svc)
	c.serviceMapping[alias] = dm
	c.serviceMapping[hostname] = dm
	c.refreshProxy()
}
