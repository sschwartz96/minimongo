package mock

import (
	"context"
	"log"
	"reflect"
	"testing"

	"github.com/sschwartz96/minimongo/db"
)

type testObj struct {
	Name  string
	Value int
}

func TestDB_Open(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
		{"0:empty collection", testDB, args{"", nil}, true},
		{"1:nil", testDB, args{"test", nil}, true},
		{"2:single object", testDB, args{"test", testObj{Name: "testObj1", Value: 123}}, false},
		{"2:single ptr object", testDB, args{"test", &testObj{Name: "testObj1", Value: 123}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Insert(tt.args.collection, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("DB.Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.object != nil {
				for _, o := range *tt.d.collectionMap[tt.args.collection] {
					if reflect.DeepEqual(tt.args.object, o) {
						return
					}
				}
				t.Errorf("could not find inserted object")
			}
		})
	}
}

func TestDB_FindOne(t *testing.T) {
	t.Parallel()
	testDB := &DB{
		collectionMap: map[string]*[]interface{}{
			"fooCollection":  {testObj{"objName", 123}},
			"foo2Collection": {testObj{"objName2", 246}},
			"fooPointerCol":  {&testObj{"objPointer", 246}},
		},
	}
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
		{
			name: "FindOne()[0]",
			args: args{
				collection: "",
				filter:     &db.Filter{"value": 123},
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: true,
		},
		{
			name: "FindOne()[1]",
			args: args{
				collection: "testCollection",
				filter:     nil,
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: true,
		},
		{
			name: "FindOne()[2]",
			args: args{
				collection: "foo2Collection",
				filter:     &db.Filter{"value": 246},
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: false,
		},
		{
			name: "FindOne()[3]",
			args: args{
				collection: "fooCollection",
				filter:     &db.Filter{"name": "objName"},
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: false,
		},
		{
			name: "FindOne()[4]",
			args: args{
				collection: "fooPointerCol",
				filter:     &db.Filter{"name": "objPointer"},
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: false,
		},
		{
			name: "FindOne()[5]",
			args: args{
				collection: "fooPointerCol",
				filter:     &db.Filter{"name": "canfindthis"},
				object:     &testObj{},
				opts:       db.CreateOptions(),
			},
			d:       testDB,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.FindOne(tt.args.collection, tt.args.object, tt.args.filter, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DB.FindOne() error = %v, wantErr %v", err, tt.wantErr)
			}

			// make sure we found what we wanted
			if !tt.wantErr {
				for _, data := range *tt.d.collectionMap[tt.args.collection] {
					dataVal := reflect.ValueOf(data)
					if dataVal.Kind() == reflect.Ptr {
						if reflect.DeepEqual(dataVal.Interface(),
							reflect.ValueOf(tt.args.object).Interface()) {
							return
						}
					} else {
						if reflect.DeepEqual(dataVal.Interface(),
							reflect.ValueOf(tt.args.object).Elem().Interface()) {
							return
						}
					}
					t.Errorf("could not find \"found\" object, ")
				}
			}
		})
	}
}

func TestDB_FindAll(t *testing.T) {
	t.Parallel()
	testDB := &DB{
		collectionMap: map[string]*[]interface{}{
			"fooCollection":  {testObj{"objName", 123}, testObj{"obj2Name", 456}, testObj{"obj3Name", 456}},
			"foo2Collection": {testObj{"objName2", 246}},
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
			name: "FindAll()[0]single",
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
		{
			name: "FindAll()[1]two values",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      &[]testObj{},
				filter:     &db.Filter{"value": 456},
				opts:       nil,
			},
			wantErr: false,
			endingSlice: &[]testObj{
				{Name: "obj2Name", Value: 456},
				{Name: "obj3Name", Value: 456},
			},
		},
		{
			name: "FindAll()[2]not pointer to slice",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      []testObj{},
				filter:     &db.Filter{"value": 456},
				opts:       nil,
			},
			wantErr:     true,
			endingSlice: &[]testObj{},
		},
		{
			name: "FindAll()[3]not slice",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      &testObj{},
				filter:     &db.Filter{"value": 456},
				opts:       nil,
			},
			wantErr:     true,
			endingSlice: &[]testObj{},
		},
		{
			name: "FindAll()[3]not slice",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      &testObj{},
				filter:     &db.Filter{},
				opts:       nil,
			},
			wantErr:     true,
			endingSlice: &[]testObj{},
		},
		{
			name: "FindAll()[4]collection does not exist",
			d:    testDB,
			args: args{
				collection: "thisIsn'tACollection",
				slice:      &[]testObj{},
				filter:     &db.Filter{},
				opts:       nil,
			},
			wantErr:     true,
			endingSlice: &[]testObj{},
		},
		{
			name: "FindAll()[5]all_fooCollection",
			d:    testDB,
			args: args{
				collection: "fooCollection",
				slice:      &[]testObj{},
				filter:     nil,
				opts:       nil,
			},
			wantErr:     false,
			endingSlice: &[]testObj{{"objName", 123}, {"obj2Name", 456}, {"obj3Name", 456}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.FindAll(tt.args.collection, tt.args.slice, tt.args.filter, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DB.FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.args.slice, tt.endingSlice) {
				t.Errorf("DB.FindAll() error = wrong slice, got: %v, wanted: %v", tt.args.slice, tt.endingSlice)
			}
		})
	}
}

func TestDB_Update(t *testing.T) {
	t.Parallel()
	testDB := &DB{
		collectionMap: map[string]*[]interface{}{
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
		{"update3_no docs",
			testDB,
			args{
				collection: "fooCollection",
				filter:     &db.Filter{"name": "notfound"},
				object:     testObj{"objNameChange", 321},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Update(tt.args.collection, tt.args.object, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				for _, data := range *tt.d.collectionMap[tt.args.collection] {
					if reflect.DeepEqual(reflect.ValueOf(data).Interface(),
						reflect.ValueOf(tt.args.object).Interface()) {
						return
					}
				}
				t.Errorf("could not find updated object")
			}
		})
	}
}

func TestDB_Upsert(t *testing.T) {
	t.Parallel()
	testDB := &DB{
		collectionMap: map[string]*[]interface{}{
			"fooCollection":  {testObj{Name: "objName", Value: 123}},
			"foo2Collection": {testObj{Name: "objName2", Value: 246}},
		},
	}
	emptyTestDB := &DB{
		collectionMap: map[string]*[]interface{}{},
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
		{"upsert0",
			emptyTestDB,
			args{
				collection: "",
				filter:     &db.Filter{"name": "objName"},
				object:     testObj{"onlyObj", 111111111},
			},
			true,
		},
		{"upsert0",
			testDB,
			args{
				collection: "",
				filter:     &db.Filter{"name": "objName"},
				object:     testObj{"objNameChange", 321},
			},
			true,
		},
		{"upsert1",
			testDB,
			args{
				collection: "fooCollection",
				filter:     nil,
				object:     testObj{"objNameChange", 321},
			},
			true,
		},
		{"upsert2",
			testDB,
			args{
				collection: "fooCollection",
				filter:     &db.Filter{"name": "objNameButNotInCollection"},
				object:     testObj{"objNameChange1", 321},
			},
			false,
		},
		{"upsert3",
			testDB,
			args{
				collection: "fooCollection",
				filter:     &db.Filter{"name": "objName"},
				object:     testObj{"objNameChange2", 321},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Upsert(tt.args.collection, tt.args.object, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				for _, data := range *tt.d.collectionMap[tt.args.collection] {
					if reflect.DeepEqual(reflect.ValueOf(data).Interface(),
						reflect.ValueOf(tt.args.object).Interface()) {
						return
					}
				}
				t.Errorf("could not find upserted object")
			}
		})
	}
}

func TestDB_Delete(t *testing.T) {
	t.Parallel()
	testDB := &DB{
		map[string]*[]interface{}{
			"fooCollection": {testObj{"obj1", 1}, testObj{"obj2", 2}, testObj{"obj3", 3}, testObj{"obj4", 4}},
		},
	}
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
		{name: "delete0", args: args{collection: ""}, d: testDB, wantErr: true},
		{name: "delete1", args: args{collection: "fooCollection", filter: nil}, d: testDB, wantErr: true},
		{name: "delete2", args: args{collection: "fooCollection", filter: &db.Filter{"name": "obj"}}, d: testDB, wantErr: true},
		{name: "delete3", args: args{collection: "fooCollection", filter: &db.Filter{"name": "obj2"}}, d: testDB, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Delete(tt.args.collection, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DB.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				for _, filterVal := range *tt.args.filter {
					for _, data := range *tt.d.collectionMap[tt.args.collection] {
						dataVal := reflect.ValueOf(data)
						log.Println("db:", *tt.d.collectionMap[tt.args.collection])
						for i := 0; i < dataVal.NumField(); i++ {
							if reflect.DeepEqual(dataVal.Field(i).Interface(),
								reflect.ValueOf(filterVal).Interface()) {
								t.Errorf("found the object still collection: %v", data)
							}
						}
					}
				}
			}
		})
	}
}

func Test_Search(t *testing.T) {
	t.Parallel()
	testDB := &DB{collectionMap: map[string]*[]interface{}{
		"testCol": {
			testObj{
				Name:  "test object 1",
				Value: 123,
			},
			testObj{
				Name:  "this is object 2",
				Value: 123,
			},
			testObj{
				Name:  "test object 3",
				Value: 123,
			},
		},
	}}
	type args struct {
		collection string
		search     string
		fields     []string
		slice      interface{}
	}
	tests := []struct {
		name        string
		d           *DB
		args        args
		wantErr     bool
		endingSlice *[]testObj
	}{
		{
			name: "Search()_0",
			args: args{
				collection: "testCol",
				fields:     []string{"Name"},
				slice:      &[]testObj{},
				search:     "object",
			},
			d:       testDB,
			wantErr: false,
			endingSlice: &[]testObj{
				{
					Name:  "test object 1",
					Value: 123,
				},
				{
					Name:  "this is object 2",
					Value: 123,
				},
				{
					Name:  "test object 3",
					Value: 123,
				},
			},
		},
		{
			name: "Search()_0",
			args: args{
				collection: "testCol",
				fields:     []string{"Name"},
				slice:      &[]testObj{},
				search:     "test",
			},
			d:       testDB,
			wantErr: false,
			endingSlice: &[]testObj{
				{
					Name:  "test object 1",
					Value: 123,
				},
				{
					Name:  "test object 3",
					Value: 123,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Search(tt.args.collection, tt.args.search, tt.args.fields, tt.args.slice); (err != nil) != tt.wantErr {
				t.Errorf("DB.Search() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.slice, tt.endingSlice) {
				t.Errorf("DB.Search() error = wrong slice, got: %v, wanted: %v", tt.args.slice, tt.endingSlice)
			}
		})
	}
}
