// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	time "time"

	istiov1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	versioned "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/application-registry/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/client/listers/istio/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// InstanceInformer provides access to a shared informer and lister for
// Instances.
type InstanceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.InstanceLister
}

type instanceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewInstanceInformer constructs a new informer for Instance type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewInstanceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredInstanceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredInstanceInformer constructs a new informer for Instance type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredInstanceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IstioV1alpha2().Instances(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IstioV1alpha2().Instances(namespace).Watch(context.TODO(), options)
			},
		},
		&istiov1alpha2.Instance{},
		resyncPeriod,
		indexers,
	)
}

func (f *instanceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredInstanceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *instanceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&istiov1alpha2.Instance{}, f.defaultInformer)
}

func (f *instanceInformer) Lister() v1alpha2.InstanceLister {
	return v1alpha2.NewInstanceLister(f.Informer().GetIndexer())
}
