// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// HTTPSourceLister helps list HTTPSources.
type HTTPSourceLister interface {
	// List lists all HTTPSources in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.HTTPSource, err error)
	// HTTPSources returns an object that can list and get HTTPSources.
	HTTPSources(namespace string) HTTPSourceNamespaceLister
	HTTPSourceListerExpansion
}

// hTTPSourceLister implements the HTTPSourceLister interface.
type hTTPSourceLister struct {
	indexer cache.Indexer
}

// NewHTTPSourceLister returns a new HTTPSourceLister.
func NewHTTPSourceLister(indexer cache.Indexer) HTTPSourceLister {
	return &hTTPSourceLister{indexer: indexer}
}

// List lists all HTTPSources in the indexer.
func (s *hTTPSourceLister) List(selector labels.Selector) (ret []*v1alpha1.HTTPSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HTTPSource))
	})
	return ret, err
}

// HTTPSources returns an object that can list and get HTTPSources.
func (s *hTTPSourceLister) HTTPSources(namespace string) HTTPSourceNamespaceLister {
	return hTTPSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// HTTPSourceNamespaceLister helps list and get HTTPSources.
type HTTPSourceNamespaceLister interface {
	// List lists all HTTPSources in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.HTTPSource, err error)
	// Get retrieves the HTTPSource from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.HTTPSource, error)
	HTTPSourceNamespaceListerExpansion
}

// hTTPSourceNamespaceLister implements the HTTPSourceNamespaceLister
// interface.
type hTTPSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all HTTPSources in the indexer for a given namespace.
func (s hTTPSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.HTTPSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HTTPSource))
	})
	return ret, err
}

// Get retrieves the HTTPSource from the indexer for a given namespace and name.
func (s hTTPSourceNamespaceLister) Get(name string) (*v1alpha1.HTTPSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("httpsource"), name)
	}
	return obj.(*v1alpha1.HTTPSource), nil
}
