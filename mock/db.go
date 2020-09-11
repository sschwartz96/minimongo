package mock

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sschwartz96/minimongo/db"
)

type DB struct {
	collectionMap map[string]([]interface{})
}

func CreateDB() *DB {
	return &DB{collectionMap: make(map[string][]interface{})}
}

func (d *DB) Open(ctx context.Context) error {
	d.collectionMap = make(map[string]([]interface{}))
	return nil
}

func (d *DB) Close(ctx context.Context) error {
	return nil
}

func (d *DB) Insert(collection string, object interface{}) error {
	if collection == "" {
		return errors.New("collection is empty")
	}
	if object == nil {
		return errors.New("object is nil")
	}

	if d.collectionMap[collection] == nil {
		d.collectionMap[collection] = make([]interface{}, 1)
		d.collectionMap[collection][0] = object
	} else {
		d.collectionMap[collection] = append(d.collectionMap[collection], object)
	}

	return nil
}

func (d *DB) FindOne(collection string, object interface{}, filter *db.Filter, opts *db.Options) error {
	if d.collectionMap[collection] == nil {
		return errors.New("collection doesn not exist")
	}

	dataSlice := d.collectionMap[collection]
	for _, data := range dataSlice {
		if compareInterfaceToFilter(data, filter) {
			return setValue(object, data)
		}
	}

	return errors.New("no object found with filter")
}

func (d *DB) FindAll(collection string, slice interface{}, filter *db.Filter, opts *db.Options) error {
	pointerVal := reflect.ValueOf(slice)
	if pointerVal.Kind() != reflect.Ptr {
		return errors.New("slice arg must be a *pointer* to slice")
	}
	sliceVal := pointerVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("slice arg does not point to a *slice*")
	}

	if d.collectionMap[collection] == nil {
		return errors.New("collection does not not exist")
	}

	if filter == nil {
		filter = &db.Filter{}
	}

	dataSlice := d.collectionMap[collection]
	for _, data := range dataSlice {
		if compareInterfaceToFilter(data, filter) {
			dataVal := reflect.ValueOf(data)
			sliceVal = reflect.Append(sliceVal, dataVal)
		}
	}

	pointerVal.Elem().Set(sliceVal)
	return nil
}

func (d *DB) Update(collection string, object interface{}, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}

	dataSlice := d.collectionMap[collection]
	for i, data := range dataSlice {
		if compareInterfaceToFilter(data, filter) {
			return setValue(&dataSlice[i], object)
		}
	}

	return errors.New("mock.DB.Update() error: no documents found")
}

func (d *DB) Upsert(collection string, object interface{}, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}
	dataSlice := d.collectionMap[collection]
	for i, data := range dataSlice {
		if compareInterfaceToFilter(data, filter) {
			return setValue(&dataSlice[i], object)
		}
	}
	return d.Insert(collection, object)
}

func (d *DB) Delete(collection string, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}
	dataSlice := d.collectionMap[collection]
	for i, data := range dataSlice {
		if compareInterfaceToFilter(data, filter) {
			// gather the slice around the object
			part1 := reflect.ValueOf(dataSlice).Slice(0, i)
			part2 := reflect.ValueOf(dataSlice).Slice(i+1, len(dataSlice))

			// create a new slice and append
			newSlice := make([]interface{}, 0)
			newSliceVal := reflect.ValueOf(newSlice)
			newSliceVal = reflect.Append(newSliceVal, part1, part2)

			// set it as the new []interface{} for that collection
			d.collectionMap[collection] = newSlice
			return nil
		}
	}
	return errors.New("mock.DB.Delete(): no documents found")
}

func (d *DB) Search(collection string, search string, fields []string, object interface{}) error {
	panic("not implemented") // TODO: Implement
}

func checkParams(collection string, filter *db.Filter) error {
	if collection == "" {
		return errors.New("collection cannot be empty")
	}
	if filter == nil || len(*filter) == 0 {
		return errors.New("filter cannot be empty or nil")
	}
	return nil
}

func compareInterfaceToFilter(a interface{}, filter *db.Filter) bool {
	aVal := reflect.ValueOf(a)

	if !aVal.IsValid() {
		return false
	}

	for filterKey, filterVal := range *filter {
		for i := 0; i < aVal.NumField(); i++ {
			fieldVal := aVal.Field(i)
			fieldName := aVal.Type().Field(i).Name
			if isLowerEqual(filterKey, fieldName) {
				if isEqual(reflect.ValueOf(filterVal), fieldVal) {
					break
				} else {
					return false
				}
			}
		}
	}
	return true
}

func isEqual(a, b reflect.Value) bool {
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		return a.Interface() == b.Interface()
	}
	return false
}

func isLowerEqual(a, b string) bool {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b)) == 0
}

func setValue(into, datafrom interface{}) error {
	if reflect.TypeOf(into).Kind() != reflect.Ptr {
		return errors.New("input object is not type pointer")
	}
	reflect.ValueOf(into).Elem().Set(reflect.ValueOf(datafrom))
	return nil
}
