package cache

import "k8s.io/client-go/tools/cache"

type SharedIndexInformers map[string]cache.SharedIndexInformer

func NewSharedIndexInformers() SharedIndexInformers {
	return make(SharedIndexInformers)
}

func (i SharedIndexInformers) Cluster(clusterName string) cache.SharedIndexInformer {
	return i[clusterName]
}

func (i SharedIndexInformers) GetStore() cache.Store {
	return i
}

func (i SharedIndexInformers) GetIndexer() cache.Indexer {
	return i
}

func (i SharedIndexInformers) Run(done <-chan struct{}) {
	for _, j := range i {
		go j.Run(done)
	}
}

func (i SharedIndexInformers) HasSynced() bool {
	for _, j := range i {
		if !j.HasSynced() {
			return false
		}
	}
	return true
}

func (i SharedIndexInformers) IndexKeys(indexedName string, indexedValue string) ([]string, error) {
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

func (i SharedIndexInformers) ByIndex(indexName string, indexedValue string) ([]interface{}, error) {
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

func (i SharedIndexInformers) Index(indexName string, obj interface{}) ([]interface{}, error) {
	panic("not implemented")
}

func (i SharedIndexInformers) ListIndexFuncValues(indexName string) []string {
	panic("not implemented")
}

func (i SharedIndexInformers) GetIndexers() cache.Indexers {
	panic("not implemented")
}

func (i SharedIndexInformers) AddIndexers(newIndexers cache.Indexers) error {
	panic("not implemented")
}

func (i SharedIndexInformers) Add(obj interface{}) error {
	panic("not implemented")
}

func (i SharedIndexInformers) Update(obj interface{}) error {
	panic("not implemented")
}

func (i SharedIndexInformers) Delete(obj interface{}) error {
	panic("not implemented")
}

func (i SharedIndexInformers) List() []interface{} {
	panic("not implemented")
}

func (i SharedIndexInformers) ListKeys() []string {
	panic("not implemented")
}

func (i SharedIndexInformers) Get(obj interface{}) (item interface{}, exists bool, err error) {
	panic("not implemented")
}

func (i SharedIndexInformers) GetByKey(key string) (item interface{}, exists bool, err error) {
	clusterName, namespace, name, err := SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, false, err
	}
	return i.Cluster(clusterName).GetStore().Get(namespace + "/" + name)
}

func (i SharedIndexInformers) Replace(i2 []interface{}, s string) error {
	panic("not implemented")
}

func (i SharedIndexInformers) Resync() error {
	panic("not implemented")
}
