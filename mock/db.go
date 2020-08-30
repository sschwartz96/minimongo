package mock

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/sschwartz96/minimongo/db"
)

type DB struct {
	collectionMap map[string]([]interface{})
}

func (d *DB) Open(ctx context.Context) error {
	d.collectionMap = make(map[string]([]interface{}))
	return nil
}

func (d *DB) Close(ctx context.Context) error {
	return nil
}

func (d *DB) Insert(collection string, object interface{}) error {
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

	dataMap := d.collectionMap[collection]
	for _, data := range dataMap {
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

	dataMap := d.collectionMap[collection]
	for _, data := range dataMap {
		if compareInterfaceToFilter(data, filter) {
			dataVal := reflect.ValueOf(data)
			sliceVal = reflect.Append(sliceVal, dataVal)
		}
	}

	pointerVal.Elem().Set(sliceVal)
	return nil
}

func (d *DB) Update(collection string, object interface{}, filter *db.Filter) error {
	panic("not implemented") // TODO: Implement
}

func (d *DB) Upsert(collection string, object interface{}, filter *db.Filter) error {
	panic("not implemented") // TODO: Implement
}

func (d *DB) Delete(collection string, filter *db.Filter) error {
	panic("not implemented") // TODO: Implement
}

func (d *DB) Search(collection string, search string, fields []string, object interface{}) error {
	panic("not implemented") // TODO: Implement
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

func setValue(object, data interface{}) error {
	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		return errors.New("input object is not type pointer")
	}
	reflect.ValueOf(object).Elem().Set(reflect.ValueOf(data))
	return nil
}
