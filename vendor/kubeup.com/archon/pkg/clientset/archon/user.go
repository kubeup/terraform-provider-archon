/*
Copyright 2016 The Archon Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package archon

import (
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	"kubeup.com/archon/pkg/cluster"
)

// UsersGetter has a method to return a UserInterface.
// A group's client should implement this interface.
type UsersGetter interface {
	Users(namespace string) UserInterface
}

// UserInterface has methods to work with User resources.
type UserInterface interface {
	Create(*cluster.User) (*cluster.User, error)
	Update(*cluster.User) (*cluster.User, error)
	Delete(name string) error
	Get(name string) (*cluster.User, error)
	List() (*cluster.UserList, error)
	Watch() (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte) (*cluster.User, error)
}

// users implements UserInterface
type users struct {
	client rest.Interface
	ns     string
}

// newUsers returns a Users
func newUsers(c *ArchonClient, namespace string) *users {
	return &users{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a user and creates it.  Returns the server's representation of the user, and an error, if there is any.
func (c *users) Create(user *cluster.User) (result *cluster.User, err error) {
	result = &cluster.User{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("users").
		Body(user).
		Do().
		Into(result)
	return
}

// Update takes the representation of a user and updates it. Returns the server's representation of the user, and an error, if there is any.
func (c *users) Update(user *cluster.User) (result *cluster.User, err error) {
	result = &cluster.User{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("users").
		Name(user.Name).
		Body(user).
		Do().
		Into(result)
	return
}

func (c *users) Patch(name string, pt types.PatchType, data []byte) (result *cluster.User, err error) {
	result = &cluster.User{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("users").
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

// Delete takes name of the user and deletes it. Returns an error if one occurs.
func (c *users) Delete(name string) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("users").
		Name(name).
		Do().
		Error()
}

// Get takes name of the user, and returns the corresponding user object, and an error if there is any.
func (c *users) Get(name string) (result *cluster.User, err error) {
	result = &cluster.User{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("users").
		Name(name).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Users that match those selectors.
func (c *users) List() (result *cluster.UserList, err error) {
	result = &cluster.UserList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("users").
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested users.
func (c *users) Watch() (watch.Interface, error) {
	return c.client.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("users").
		Watch()
}
