package main

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
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
	rateLimitingQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	// 4. 注册对应事件的处理函数
	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// do something
			fmt.Println("Add Event")
			// 将 obj 的 key 传入 queue中
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				print(err)
			}
			rateLimitingQueue.AddRateLimited(key)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// do something
			fmt.Print("Update Event")
			// 将 new obj 的 key 传入队列中
			key, _ := cache.MetaNamespaceKeyFunc(newObj)
			rateLimitingQueue.AddRateLimited(key)
		},
		DeleteFunc: func(obj interface{}) {
			// do something
			fmt.Println("Delete Event")
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			rateLimitingQueue.AddRateLimited(key)
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
