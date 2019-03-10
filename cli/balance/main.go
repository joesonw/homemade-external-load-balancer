package main

import (
	"cli/balance/app"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"pkg/ips"
	"pkg/nameservers"
	"pkg/proxy"
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

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
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

	p, err := proxy.New(config.Providers.Proxy)
	if err != nil {
		log.Printf("unable to initiate proxy provider: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("proxy inited \n")

	client := app.New(config, k8sClient, nameserver, ip, p)
	err = client.Start()
	if err != nil {
		log.Printf("unable to start load balancer: %s\n", err.Error())
		os.Exit(1)
	}
}
