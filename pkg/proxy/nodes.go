package proxy

import (
	"time"

	"math/rand"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (s *Server) startWatchNodes() {
	resync := time.Second * 600
	informerFactory := informers.NewSharedInformerFactory(s.clientset, resync)
	nodeInformer := informerFactory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				node := obj.(*corev1.Node)
				s.handleAddNode(node)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldNode := oldObj.(*corev1.Node)
				newNode := newObj.(*corev1.Node)
				s.handleDeleteNode(oldNode)
				s.handleAddNode(newNode)
			},
			DeleteFunc: func(obj interface{}) {
				node := obj.(*corev1.Node)
				s.handleDeleteNode(node)
			},
		},
		resync,
	)

	informerFactory.Start(wait.NeverStop)
}

func (s *Server) handleAddNode(node *corev1.Node) {
	s.nodeLock.Lock()
	defer s.nodeLock.Unlock()
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			s.nodes[node.ObjectMeta.Name] = addr.Address
		}
	}

}

func (s *Server) handleDeleteNode(node *corev1.Node) {
	s.nodeLock.Lock()
	defer s.nodeLock.Unlock()
	delete(s.nodes, node.ObjectMeta.Name)
}

func (s *Server) getRandomNodeIP() string {
	s.nodeLock.RLock()
	defer s.nodeLock.RUnlock()
	length := len(s.nodes)
	if length == 0 {
		return ""
	}
	index := rand.Intn(length)
	i := 0
	for _, ip := range s.nodes {
		if i == index {
			return ip
		}
		i++
	}
	return ""
}
