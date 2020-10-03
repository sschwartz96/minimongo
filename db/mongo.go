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
	if opts.Limit > 0 {
		o.SetLimit(opts.Limit)
	}
	if opts.Skip > 0 {
		o.SetSkip(opts.Skip)
	}
	if opts.Sort != nil {
		o.SetSort(bson.M{opts.Sort.Key: opts.Sort.Value})
	}
	return o
}

// convertToMongoOne converts database.Options to options.FindOneOptions
func ConvertToFindOneOptions(opts *Options) *options.FindOneOptions {
	if opts == nil {
		return options.FindOne()
	}
	o := options.FindOne()
	if opts.Skip > 0 {
		o.SetSkip(opts.Skip)
	}
	if opts.Sort != nil {
		o.SetSort(bson.M{opts.Sort.Key: opts.Sort.Value})
	}
	return o
}
