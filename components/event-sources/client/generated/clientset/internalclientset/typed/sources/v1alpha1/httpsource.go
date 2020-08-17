// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	scheme "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HTTPSourcesGetter has a method to return a HTTPSourceInterface.
// A group's client should implement this interface.
type HTTPSourcesGetter interface {
	HTTPSources(namespace string) HTTPSourceInterface
}

// HTTPSourceInterface has methods to work with HTTPSource resources.
type HTTPSourceInterface interface {
	Create(*v1alpha1.HTTPSource) (*v1alpha1.HTTPSource, error)
	Update(*v1alpha1.HTTPSource) (*v1alpha1.HTTPSource, error)
	UpdateStatus(*v1alpha1.HTTPSource) (*v1alpha1.HTTPSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.HTTPSource, error)
	List(opts v1.ListOptions) (*v1alpha1.HTTPSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HTTPSource, err error)
	HTTPSourceExpansion
}

// hTTPSources implements HTTPSourceInterface
type hTTPSources struct {
	client rest.Interface
	ns     string
}

// newHTTPSources returns a HTTPSources
func newHTTPSources(c *SourcesV1alpha1Client, namespace string) *hTTPSources {
	return &hTTPSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the hTTPSource, and returns the corresponding hTTPSource object, and an error if there is any.
func (c *hTTPSources) Get(name string, options v1.GetOptions) (result *v1alpha1.HTTPSource, err error) {
	result = &v1alpha1.HTTPSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpsources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HTTPSources that match those selectors.
func (c *hTTPSources) List(opts v1.ListOptions) (result *v1alpha1.HTTPSourceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.HTTPSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hTTPSources.
func (c *hTTPSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("httpsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a hTTPSource and creates it.  Returns the server's representation of the hTTPSource, and an error, if there is any.
func (c *hTTPSources) Create(hTTPSource *v1alpha1.HTTPSource) (result *v1alpha1.HTTPSource, err error) {
	result = &v1alpha1.HTTPSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("httpsources").
		Body(hTTPSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a hTTPSource and updates it. Returns the server's representation of the hTTPSource, and an error, if there is any.
func (c *hTTPSources) Update(hTTPSource *v1alpha1.HTTPSource) (result *v1alpha1.HTTPSource, err error) {
	result = &v1alpha1.HTTPSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpsources").
		Name(hTTPSource.Name).
		Body(hTTPSource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *hTTPSources) UpdateStatus(hTTPSource *v1alpha1.HTTPSource) (result *v1alpha1.HTTPSource, err error) {
	result = &v1alpha1.HTTPSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpsources").
		Name(hTTPSource.Name).
		SubResource("status").
		Body(hTTPSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the hTTPSource and deletes it. Returns an error if one occurs.
func (c *hTTPSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpsources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hTTPSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpsources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched hTTPSource.
func (c *hTTPSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HTTPSource, err error) {
	result = &v1alpha1.HTTPSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("httpsources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
