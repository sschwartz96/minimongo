package db

import "context"

// Database defines database functionality
type Database interface {
	Open(ctx context.Context) error
	Close(ctx context.Context) error

	Insert(collection string, object interface{}) error
	FindOne(collection string, object interface{}, filter *Filter, opts *Options) error
	FindAll(collection string, object interface{}, filter *Filter, opts *Options) error
	Update(collection string, object interface{}, filter *Filter) error
	Upsert(collection string, object interface{}, filter *Filter) error
	Delete(collection string, filter *Filter) error
	Search(collection, search string, fields []string, object interface{}) error
}

type Filter map[string]interface{}

type Options struct {
	limit int64
	skip  int64
	sort  *sortOption
}

type sortOption struct {
	key   string
	value int // 1 ascending, -1 descending
}

func CreateOptions() *Options {
	return &Options{}
}

func (o *Options) SetLimit(v int64) *Options {
	o.limit = v
	return o
}

func (o *Options) SetSkip(v int64) *Options {
	o.skip = v
	return o
}

// SetSort sets the sort of the returned documents
// + value = ascending, - value = descending
func (o *Options) SetSort(key string, value int) *Options {
	if value > 0 {
		value = 1
	} else {
		value = -1
	}
	o.sort = &sortOption{key: key, value: value}
	return o
}
