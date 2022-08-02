package main

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 1. 初始化一个config, 使用默认的配置文件, 这里基本上和RESTClient一致
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	// 2. 使用kubernetes包下的NewforConfig, 就不需要我们配置GroupVersion, apiPath 之类的东西
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// 使用clientSet初始化一个CoreV1下的PodsClient
	podsClient := clientSet.CoreV1().Pods("default")
	pod, err := podsClient.Get(context.TODO(), "test-pod", v1.GetOptions{})
	if err != nil {
		print(err)
	} else {
		print(pod.Name)
	}
}
