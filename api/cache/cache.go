package cache

import (
	"k8s.io/client-go/tools/cache"
	"time"
)

type SharedIndexInformer interface {
	cache.SharedIndexInformer
	cache.Store
	cache.Indexer
	Cluster(name string) cache.SharedIndexInformer
}

type impl map[string]cache.SharedIndexInformer

func NewSharedIndexInformers(informers map[string]cache.SharedIndexInformer) SharedIndexInformer {
	return impl(informers)
}

func (i impl) Cluster(name string) cache.SharedIndexInformer {
	return i[name]
}

func (i impl) GetStore() cache.Store {
	return i
}

func (i impl) GetIndexer() cache.Indexer {
	return i
}

func (i impl) Run(done <-chan struct{}) {
	for _, j := range i {
		go j.Run(done)
	}
}

func (i impl) HasSynced() bool {
	for _, j := range i {
		if !j.HasSynced() {
			return false
		}
	}
	return true
}

func (i impl) IndexKeys(indexedName string, indexedValue string) ([]string, error) {
	var keys []string
	for _, j := range i {
		v, err := j.GetIndexer().IndexKeys(indexedName, indexedValue)
		if err != nil {
			return nil, err
		}
		keys = append(keys, v...)
	}
	return keys, nil
}

func (i impl) ByIndex(indexName string, indexedValue string) ([]interface{}, error) {
	var byIndex []interface{}
	for _, j := range i {
		v, err := j.GetIndexer().ByIndex(indexName, indexedValue)
		if err != nil {
			return nil, err
		}
		byIndex = append(byIndex, v...)

	}
	return byIndex, nil
}

func (i impl) Index(indexName string, obj interface{}) ([]interface{}, error) {
	panic("not implemented")
}

func (i impl) ListIndexFuncValues(indexName string) []string {
	panic("not implemented")
}

func (i impl) GetIndexers() cache.Indexers {
	panic("not implemented")
}

func (i impl) AddIndexers(newIndexers cache.Indexers) error {
	panic("not implemented")
}

func (i impl) Add(obj interface{}) error {
	panic("not implemented")
}

func (i impl) Update(obj interface{}) error {
	panic("not implemented")
}

func (i impl) Delete(obj interface{}) error {
	panic("not implemented")
}

func (i impl) List() []interface{} {
	panic("not implemented")
}

func (i impl) ListKeys() []string {
	panic("not implemented")
}

func (i impl) Get(obj interface{}) (item interface{}, exists bool, err error) {
	panic("not implemented")
}

func (i impl) GetByKey(key string) (item interface{}, exists bool, err error) {
	cluster, namespace, name, err := SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, false, err
	}
	return i.Cluster(cluster).GetStore().Get(namespace + "/" + name)
}

func (i impl) Replace(i2 []interface{}, s string) error {
	panic("not implemented")
}

func (i impl) Resync() error {
	panic("not implemented")
}

func (i impl) AddEventHandler(handler cache.ResourceEventHandler) {
	for _, j := range i {
		j.AddEventHandler(handler)
	}
}

func (i impl) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	panic("implement me")
}

func (i impl) GetController() cache.Controller {
	panic("implement me")
}

func (i impl) LastSyncResourceVersion() string {
	panic("implement me")
}

func (i impl) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	panic("implement me")
}
