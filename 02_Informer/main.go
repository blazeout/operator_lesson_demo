package main

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 1. 创建配置文件 config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	// 2. 创建client-set
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// 3. 根据具体的client创建具体资源对象的informer
	//factory := informers.NewSharedInformerFactory(clientSet, 0)
	// 如何指定具体的namespace
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, 0, informers.WithNamespace("default"))
	informer := factory.Core().V1().Pods().Informer()
	// 4. 注册对应事件的处理函数
	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// do something
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// do something
		},
		DeleteFunc: func(obj interface{}) {
			// do something
		},
	}
	informer.AddEventHandler(funcs)

	// 5. 启动informer
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	// 6. 等待informer同步完成
	factory.WaitForCacheSync(stopCh)
	<-stopCh
}
