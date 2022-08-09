package pkg

import (
	"context"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreInformer "k8s.io/client-go/informers/core/v1"
	netWorkInformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	coreLister "k8s.io/client-go/listers/core/v1"
	netWorkLister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"time"
)

const workerNum = 5
const MaxRetry = 10

// Event Handler 处理完事件之后会向work queue里面插入数据供给Worker消费
type customController struct {
	client        kubernetes.Interface
	serviceLister coreLister.ServiceLister
	ingressLister netWorkLister.IngressLister
	queue         workqueue.RateLimitingInterface
}

func (c *customController) addServiceFunc(obj interface{}) {
	// 处理完成之后需要将obj传入queue中
	c.enQueue(obj)
}

func (c *customController) updateServiceFunc(oldObj interface{}, newObj interface{}) {
	//todo: 比较两个对象的annotation是否相同
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}

}

func (c *customController) enQueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)
}

func (c *customController) deleteIngressFunc(obj interface{}) {
	ingress := obj.(*v1.Ingress)
	ownerReference := v12.GetControllerOf(ingress)
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	c.enQueue(ingress.Namespace + "/" + ingress.Name)
}

// Run 方法里面写Worker的处理逻辑
func (c *customController) Run(stopCh chan struct{}) {
	// 开启5个goroutine 来调用我们的worker方法
	for i := 0; i < workerNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func (c *customController) worker() {
	for c.processNextItem() {

	}
}

func (c *customController) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	// 做完处理之后需要移除item
	defer c.queue.Done(item)
	key := item.(string)
	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

func (c *customController) syncService(item string) error {
	// 首先获取namespace 和 name
	namespace, name, err := cache.SplitMetaNamespaceKey(item)
	if err != nil {
		return err
	}

	// 删除 获取 Service是否存在即可
	service, err := c.serviceLister.Services(namespace).Get(name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// 新增 和 更新 Service 判断 annotation的情况 和 ingress 是否存在
	_, ok := service.GetAnnotations()["ingress/http"]
	// 获取ingress
	ingress, err := c.ingressLister.Ingresses(namespace).Get(name)
	if err != nil {
		return err
	}
	if ok && errors.IsNotFound(err) {
		// 创建 ingress
		// 通过client与 api-Server通信
		ig := c.createIngress(service)
		_, err := c.client.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ig, v12.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		// 删除ingress
		err := c.client.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, v12.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *customController) createIngress(service *v13.Service) *v1.Ingress {
	ingress := v1.Ingress{}
	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := v1.PathTypePrefix
	ingress.OwnerReferences = []v12.OwnerReference{
		*v12.NewControllerRef(service, v1.SchemeGroupVersion.WithKind("Service")),
	}
	ingress.Spec = v1.IngressSpec{
		Rules: []v1.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: service.Name,
										Port: v1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingress
}

func (c *customController) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) <= MaxRetry {
		c.queue.AddRateLimited(key)
	}
	runtime.HandleError(err)
	c.queue.Forget(key)
}

func NewCustomController(client kubernetes.Interface, serviceInformer coreInformer.ServiceInformer, ingressInformer netWorkInformer.IngressInformer) customController {
	controller := customController{
		client:        client,
		serviceLister: serviceInformer.Lister(),
		ingressLister: ingressInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "IngressManager"),
	}
	// 增加事件处理函数
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addServiceFunc,
		UpdateFunc: controller.updateServiceFunc,
	})
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controller.deleteIngressFunc,
	})

	return controller
}
