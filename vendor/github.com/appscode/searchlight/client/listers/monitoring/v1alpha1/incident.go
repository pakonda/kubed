/*
Copyright 2018 The Searchlight Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/appscode/searchlight/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// IncidentLister helps list Incidents.
type IncidentLister interface {
	// List lists all Incidents in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Incident, err error)
	// Incidents returns an object that can list and get Incidents.
	Incidents(namespace string) IncidentNamespaceLister
	IncidentListerExpansion
}

// incidentLister implements the IncidentLister interface.
type incidentLister struct {
	indexer cache.Indexer
}

// NewIncidentLister returns a new IncidentLister.
func NewIncidentLister(indexer cache.Indexer) IncidentLister {
	return &incidentLister{indexer: indexer}
}

// List lists all Incidents in the indexer.
func (s *incidentLister) List(selector labels.Selector) (ret []*v1alpha1.Incident, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Incident))
	})
	return ret, err
}

// Incidents returns an object that can list and get Incidents.
func (s *incidentLister) Incidents(namespace string) IncidentNamespaceLister {
	return incidentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// IncidentNamespaceLister helps list and get Incidents.
type IncidentNamespaceLister interface {
	// List lists all Incidents in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Incident, err error)
	// Get retrieves the Incident from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Incident, error)
	IncidentNamespaceListerExpansion
}

// incidentNamespaceLister implements the IncidentNamespaceLister
// interface.
type incidentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Incidents in the indexer for a given namespace.
func (s incidentNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Incident, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Incident))
	})
	return ret, err
}

// Get retrieves the Incident from the indexer for a given namespace and name.
func (s incidentNamespaceLister) Get(name string) (*v1alpha1.Incident, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("incident"), name)
	}
	return obj.(*v1alpha1.Incident), nil
}
