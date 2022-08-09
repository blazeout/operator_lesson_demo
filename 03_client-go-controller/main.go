package main

import (
	"github.com/blazeout/operator_lesson_demo/03_client-go-controller/pkg"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

// 实现一个自定义的controller, 负责监控Service对象的变更
// 根据Service对象的增加或者删除或者更新. 来决定我们ingress的变化 根据Service的 annotation ingress/http: true
func main() {
	// 1. 需要一个config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln("can not get config file err:", err)
			return
		}
		config = inClusterConfig
	}
	//2. 生成我们的clientSet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("can not create clientSet err: ", err)
	}
	// 3. 创建对应监控资源类型的Informer, Service Informer 和 Ingress Informer
	// 使用工厂方法创建 factory
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))
	serviceInformer := informerFactory.Core().V1().Services()
	ingressInformer := informerFactory.Networking().V1().Ingresses()

	// 4. 注册对应的 Event Handler 交由 newController方法完成
	customController := pkg.NewCustomController(clientset, serviceInformer, ingressInformer)

	// 5. informer.Start
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	customController.Run(stopCh)
}
