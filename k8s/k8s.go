package k8s

import (
	"errors"
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const namespaceSecret = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

var (
	ClientBuilder func() (kubernetes.Interface, error) = concretBuilder
	ReadFile      func(string) ([]byte, error)         = ioutil.ReadFile
)

func concretBuilder() (kubernetes.Interface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(k8sConfig)
}

func GetConfigMapData(namespace, configmap string) (map[string]string, error) {
	kc, err := ClientBuilder()
	if err != nil {
		return nil, err
	}
	cm, err := kc.CoreV1().ConfigMaps(namespace).Get(configmap, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func WatchConfigMap(namespace string) (<-chan error, error) {
	kc, err := ClientBuilder()
	if err != nil {
		return nil, err
	}
	watcher, err := kc.CoreV1().ConfigMaps(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	c := make(chan error)
	go func() {
		defer close(c)
		for e := range watcher.ResultChan() {
			if e.Type == watch.Error {
				c <- errors.New("error reading configmap")
			} else {
				c <- nil
			}
		}
	}()
	return c, nil
}

func GetCurrentNamespace() (string, error) {
	ns, err := ReadFile(namespaceSecret)
	return string(ns), err
}
