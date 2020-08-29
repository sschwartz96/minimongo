package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// convertToMongoFilter converts database.Filter to a bson.M document
func ConvertToMongoFilter(filter *Filter) bson.M {
	if filter == nil {
		return bson.M{}
	}
	return bson.M(*filter)
}

// convertToFindOptions converts database.Options to options.FindOptions
func ConvertToFindOptions(opts *Options) *options.FindOptions {
	if opts == nil {
		return options.Find()
	}
	o := options.Find()
	if opts.limit > 0 {
		o.SetLimit(opts.limit)
	}
	if opts.skip > 0 {
		o.SetSkip(opts.skip)
	}
	if opts.sort != nil {
		o.SetSort(bson.M{opts.sort.key: opts.sort.value})
	}
	return o
}

// convertToMongoOne converts database.Options to options.FindOneOptions
func ConvertToFindOneOptions(opts *Options) *options.FindOneOptions {
	if opts == nil {
		return options.FindOne()
	}
	o := options.FindOne()
	if opts.skip > 0 {
		o.SetSkip(opts.skip)
	}
	if opts.sort != nil {
		o.SetSort(bson.M{opts.sort.key: opts.sort.value})
	}
	return o
}
