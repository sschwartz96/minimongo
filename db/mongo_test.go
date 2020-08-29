package db

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestConvertToMongoFilter(t *testing.T) {
	type args struct {
		filter *Filter
	}
	tests := []struct {
		name string
		args args
		want bson.M
	}{
		{"0:nil", args{nil}, bson.M{}},
		{"1:empty", args{&Filter{}}, bson.M{}},
		{"2:1 element", args{&Filter{"foo": "bar"}}, bson.M{"foo": "bar"}},
		{"3:2 elements", args{&Filter{"foo": "bar", "integer": 123}}, bson.M{"foo": "bar", "integer": 123}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToMongoFilter(tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToMongoFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToFindOptions(t *testing.T) {
	type args struct {
		opts *Options
	}
	tests := []struct {
		name string
		args args
		want *options.FindOptions
	}{
		{"nil", args{opts: nil}, options.Find()},
		{"limit", args{opts: CreateOptions().SetLimit(123)}, options.Find().SetLimit(123)},
		{"skip", args{opts: CreateOptions().SetSkip(123)}, options.Find().SetSkip(123)},
		{"sort", args{opts: CreateOptions().SetSort("date", -1)}, options.Find().SetSort(bson.M{"date": -1})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToFindOptions(tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToFindOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToFindOneOptions(t *testing.T) {
	type args struct {
		opts *Options
	}
	tests := []struct {
		name string
		args args
		want *options.FindOneOptions
	}{
		{"nil", args{opts: nil}, options.FindOne()},
		{"limit", args{opts: CreateOptions().SetLimit(123)}, options.FindOne()},
		{"skip", args{opts: CreateOptions().SetSkip(123)}, options.FindOne().SetSkip(123)},
		{"sort", args{opts: CreateOptions().SetSort("date", -1)}, options.FindOne().SetSort(bson.M{"date": -1})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToFindOneOptions(tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToFindOneOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
