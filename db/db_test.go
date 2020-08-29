package db

import (
	"reflect"
	"testing"
)

func TestCreateOptions(t *testing.T) {
	tests := []struct {
		name string
		want *Options
	}{
		{"create options", &Options{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_SetLimit(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name string
		o    *Options
		args args
		want *Options
	}{
		{"0:nil", CreateOptions(), args{10}, &Options{limit: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.SetLimit(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.SetLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_SetSkip(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name string
		o    *Options
		args args
		want *Options
	}{
		{"0:nil", CreateOptions(), args{10}, &Options{skip: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.SetSkip(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Options.SetSkip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_SetSort(t *testing.T) {
	type args struct {
		key   string
		value int
	}
	tests := []struct {
		name string
		o    *Options
		args args
		want *Options
	}{
		{"0: zero value", CreateOptions(), args{"foo", 0}, &Options{sort: &sortOption{"foo", -1}}},
		{"1: one value", CreateOptions(), args{"foo", 1}, &Options{sort: &sortOption{"foo", 1}}},
		{"-1: negative one value", CreateOptions(), args{"foo", -1}, &Options{sort: &sortOption{"foo", -1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.SetSort(tt.args.key, tt.args.value); !reflect.DeepEqual(*(got.sort), *(tt.want.sort)) {
				t.Errorf("Options.SetSort() = %v, want %v", *(got.sort), *(tt.want.sort))
			}
		})
	}
}
