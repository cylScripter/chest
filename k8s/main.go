package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

const kubeConfigFilePath = "./config"

type K8sConfig struct {
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

// 读取kubeconfig 配置文件
func (this *K8sConfig) K8sRestConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)

	if err != nil {
		log.Fatal(err)
	}

	return config
}

// 初始化 clientSet
func (this *K8sConfig) InitClient() *kubernetes.Clientset {
	c, err := kubernetes.NewForConfig(this.K8sRestConfig())

	if err != nil {
		log.Fatal(err)
	}

	return c
}

// 初始化 dynamicClient
func (this *K8sConfig) InitDynamicClient() dynamic.Interface {
	c, err := dynamic.NewForConfig(this.K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}

	return c
}

// 初始化 DiscoveryClient
func (this *K8sConfig) InitDiscoveryClient() *discovery.DiscoveryClient {
	return discovery.NewDiscoveryClient(this.InitClient().RESTClient())
}

func main() {
	// 使用的是上文提到的配置加载对象
	cliset := NewK8sConfig().InitClient()

	configMaps, err := cliset.CoreV1().ConfigMaps("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, cm := range configMaps.Items {

		fmt.Printf("configName: %v \n", cm.Name)

	}

	// 获取default命名空间下的所有POD
	podsList, err := cliset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	for _, pod := range podsList.Items {
		fmt.Printf("podName: %v \n", pod)
	}

}
