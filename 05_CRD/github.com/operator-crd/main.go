package main

import (
	"github.com/operator_lesson_demo/05_CRD/github.com/opeator-crd/pkg/generated/clientset/versioned"
	"github.com/operator_lesson_demo/05_CRD/github.com/opeator-crd/pkg/generated/informers/externalversions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	list, err := clientset.CrdV1().Foos("default").List(nil)
	if err != nil {
		panic(err)
	}
	for _, foo := range list.Items {
		println(foo.Name)
	}

	factory := externalversions.NewSharedInformerFactory(clientset, 0)
	factory.Crd().V1().Foos().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//todo
		},
	})
	//TODO
}
