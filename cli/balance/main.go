package main

import (
	"cli/balance/app"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"pkg/ips"
	"pkg/nameservers"
	"pkg/proxy"

	"crypto/tls"

	"pkg/dns"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	configBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Printf("unable to read config %s: %s\n", os.Args[1], err.Error())
		os.Exit(1)
	}

	config := &app.Config{}
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		log.Printf("unable to parse config file: %s\n", err.Error())
		os.Exit(1)
	}

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	}

	if err != nil {
		log.Printf("unable to find kubernetes config: %s\n", err.Error())
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Printf("unable to connect to cluster: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("kubernetes inited \n")

	nameserver, err := nameservers.New(config.Providers.Nameserver)
	if err != nil {
		log.Printf("unable to initiate nameserver provider: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("nameserver inited \n")

	ip, err := ips.New(config.Providers.IP)
	if err != nil {
		log.Printf("unable to initiate nameserver provider: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("ip inited \n")

	var cert tls.Certificate
	if config.TLS != nil {
		cert, err = tls.LoadX509KeyPair(config.TLS.Cert, config.TLS.Key)
		if err != nil {
			log.Printf("unable to read tls cert/key: %s\n", err.Error())
			os.Exit(1)
		}
	}

	proxyServer := proxy.New(clientset, &cert, config.Proxy.Host)
	err = proxyServer.Start()
	if err != nil {
		log.Printf("unable to start proxy server: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("proxy inited \n")

	dnsServer := dns.New(config.Domain, config.SubDomain, config.DNS.Host, config.DNS.Port, ip)
	go func() {
		err = dnsServer.StartUDP()
		if err != nil {
			log.Printf("unable to srtart dns udp: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	go func() {
		err = dnsServer.StartTCP()
		if err != nil {
			log.Printf("unable to srtart dns tcp: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	client := app.New(config, clientset, nameserver, dnsServer, ip, proxyServer)
	err = client.Start()
	if err != nil {
		log.Printf("unable to start load balancer: %s\n", err.Error())
		os.Exit(1)
	}

	<-wait.NeverStop
}
