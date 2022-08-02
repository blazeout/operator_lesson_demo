package main

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 1. 初始化一个RESTClient的config, 使用默认的配置文件
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	// 2. 根据rest.RESTClientFor方法中要求 groupVersion 和 NegotiatedSerializer 不能为空 所以要设置
	config.APIPath = "/api"
	config.GroupVersion = &v1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	// 3. 创建Pod对象并使用RESTClient发起对api-server的httpGet请求 获取对应Pod信息
	pod := v1.Pod{}
	err = restClient.Get().Namespace("default").Resource("pods").Name("test-pod").Do(context.TODO()).Into(&pod)
	if err != nil {
		print(err)
	} else {
		print(pod.Name)
	}
}
