package mock

import (
	"context"
	"reflect"
	"testing"

	"github.com/sschwartz96/minimongo/db"
)

type testObj struct {
	Name  string
	Value int
}

func TestDB_Open(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		{
			"open",
			&DB{},
			args{
				ctx: context.Background(),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Open(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DB.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Close(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		{"close", &DB{}, args{ctx: context.Background()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DB.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Insert(t *testing.T) {
	testDB := &DB{}
	err := testDB.Open(context.Background())
	if err != nil {
		t.Fatal("error opening db:", err)
	}

	type args struct {
		collection string
		object     interface{}
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		{"0:nil test", testDB, args{"test", nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Insert(tt.args.collection, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("DB.Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_FindOne(t *testing.T) {
	type args struct {
		collection string
		object     interface{}
		filter     *db.Filter
		opts       *db.Options
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.FindOne(tt.args.collection, tt.args.object, tt.args.filter, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DB.FindOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_FindAll(t *testing.T) {
	testDB := &DB{
		collectionMap: map[string][]interface{}{
			"fooCollection":  {testObj{Name: "objName", Value: 123}},
			"foo2Collection": {testObj{Name: "objName2", Value: 246}},
		},
	}
	type args struct {
		collection string
		slice      interface{}
		filter     *db.Filter
		opts       *db.Options
	}
	tests := []struct {
		name        string
		d           *DB
		args        args
		wantErr     bool
		endingSlice *[]testObj
	}{
		{
			name: "findall",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      &[]testObj{},
				filter:     &db.Filter{"name": "objName"},
				opts:       nil,
			},
			wantErr: false,
			endingSlice: &[]testObj{
				{Name: "objName", Value: 123},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.FindAll(tt.args.collection, tt.args.slice, tt.args.filter, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DB.FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.slice, tt.endingSlice) {
				t.Errorf("DB.FindAll() error = wrong slice, got: %v, wanted: %v", tt.args.slice, tt.endingSlice)
			}
		})
	}
}

func TestDB_Update(t *testing.T) {
	testDB := &DB{
		collectionMap: map[string][]interface{}{
			"fooCollection":  {testObj{Name: "objName", Value: 123}},
			"foo2Collection": {testObj{Name: "objName2", Value: 246}},
		},
	}
	type args struct {
		collection string
		object     interface{}
		filter     *db.Filter
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		{"update0",
			testDB,
			args{
				collection: "",
				filter:     &db.Filter{"name": "objName"},
				object:     testObj{"objNameChange", 321},
			},
			true,
		},
		{"update1",
			testDB,
			args{
				collection: "fooCollection",
				filter:     nil,
				object:     testObj{"objNameChange", 321},
			},
			true,
		},
		{"update2",
			testDB,
			args{
				collection: "fooCollection",
				filter:     &db.Filter{"name": "objName"},
				object:     testObj{"objNameChange", 321},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Update(tt.args.collection, tt.args.object, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				for _, data := range tt.d.collectionMap[tt.args.collection] {
					if reflect.DeepEqual(reflect.ValueOf(data).Interface(),
						reflect.ValueOf(tt.args.object).Interface()) {
						return
					}
					t.Errorf("could not find upated object")
				}
			}
		})
	}
}

func TestDB_Upsert(t *testing.T) {
	type args struct {
		collection string
		object     interface{}
		filter     *db.Filter
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Upsert(tt.args.collection, tt.args.object, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Delete(t *testing.T) {
	type args struct {
		collection string
		filter     *db.Filter
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Delete(tt.args.collection, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Search(t *testing.T) {
	type args struct {
		collection string
		search     string
		fields     []string
		object     interface{}
	}
	tests := []struct {
		name    string
		d       *DB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Search(tt.args.collection, tt.args.search, tt.args.fields, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("DB.Search() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
