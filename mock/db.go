package mock

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/sschwartz96/stockpile/db"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DB struct {
	collectionMap map[string](*[]interface{})
}

func CreateDB() *DB {
	return &DB{collectionMap: make(map[string]*[]interface{})}
}

func (d *DB) Open(ctx context.Context) error {
	d.collectionMap = make(map[string](*[]interface{}))
	return nil
}

func (d *DB) Close(ctx context.Context) error {
	d.collectionMap = nil
	return nil
}

func (d *DB) Insert(collection string, object interface{}) error {
	if collection == "" {
		return errors.New("collection is empty")
	}
	if object == nil {
		return errors.New("object is nil")
	}

	// this allows pointers to be derefenced
	objVal := reflect.ValueOf(object)
	toInsert := objVal.Interface()
	if objVal.Kind() == reflect.Ptr {
		toInsert = objVal.Elem().Interface()
	}

	if d.collectionMap[collection] == nil {
		col := make([]interface{}, 1)
		col[0] = toInsert
		d.collectionMap[collection] = &col
	} else {
		col := d.collectionMap[collection]
		*col = append(*col, toInsert)
	}

	return nil
}

func (d *DB) FindOne(collection string, object interface{}, filter *db.Filter, opts *db.Options) error {
	if d.collectionMap[collection] == nil {
		return errors.New("collection doesn not exist")
	}

	// grab the value of the object which should be a ptr
	pointerVal := reflect.ValueOf(object)
	if pointerVal.Kind() != reflect.Ptr {
		return errors.New("object arg must be a *pointer* (to [Type])")
	}
	// get the slice type of our object type
	sliceType := reflect.SliceOf(pointerVal.Elem().Type())
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)

	err := d.findAll(collection, &sliceVal, filter, opts)
	if err != nil {
		return fmt.Errorf("error finding objects: %v", err)
	}

	if sliceVal.Len() == 0 {
		return fmt.Errorf("no object found based on filter")
	}

	return setValue(object, sliceVal.Index(0).Interface())
}

func (d *DB) FindAll(collection string, slice interface{}, filter *db.Filter, opts *db.Options) error {
	pointerVal := reflect.ValueOf(slice)
	if pointerVal.Kind() != reflect.Ptr {
		return errors.New("slice arg must be a *pointer* (to slice)")
	}
	sliceVal := pointerVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("slice arg does not point to a *slice*")
	}

	err := d.findAll(collection, &sliceVal, filter, opts)
	if err != nil {
		return fmt.Errorf("error in finding: %v", err)
	}

	pointerVal.Elem().Set(sliceVal)
	return nil
}

func (d *DB) findAll(collection string, sliceVal *reflect.Value, filter *db.Filter, opts *db.Options) error {
	if d.collectionMap[collection] == nil {
		return errors.New("collection does not not exist")
	}

	if filter == nil {
		filter = &db.Filter{}
	}

	if opts == nil {
		opts = db.CreateOptions()
	}

	var limitCounter int64
	dataSlice := *d.collectionMap[collection]

	for i := opts.Skip; int(i) < len(dataSlice); i++ {
		data := dataSlice[i]
		if compareInterfaceToFilter(data, filter) {
			appendSliceVal(sliceVal, data)

			limitCounter++
			if limitCounter == opts.Limit {
				break
			}
		}
	}

	// sort the data
	if opts.Sort != nil && opts.Sort.Value != 0 {
		sortSlice(sliceVal, opts.Sort)
	}

	return nil
}

func sortSlice(sliceVal *reflect.Value, sortOpt *db.SortOption) *reflect.Value {
	sort.SliceStable(sliceVal.Interface(), generateLessFunc(sliceVal, sortOpt))
	return sliceVal
}

func generateLessFunc(sliceVal *reflect.Value, sortOpt *db.SortOption) func(i, j int) bool {
	return func(i, j int) bool {
		var iVal, jVal reflect.Value
		if sliceVal.Index(i).Kind() == reflect.Ptr {
			iVal = sliceVal.Index(i).Elem().FieldByNameFunc(matchFieldFunc(sortOpt.Key))
			jVal = sliceVal.Index(j).Elem().FieldByNameFunc(matchFieldFunc(sortOpt.Key))
		} else {
			iVal = sliceVal.Index(i).FieldByNameFunc(matchFieldFunc(sortOpt.Key))
			jVal = sliceVal.Index(j).FieldByNameFunc(matchFieldFunc(sortOpt.Key))
		}
		switch iVal.Kind() {

		case reflect.Int, reflect.Int64, reflect.Int32:
			if sortOpt.Value > 0 {
				return iVal.Int() < jVal.Int()
			}
			return iVal.Int() > jVal.Int()

		case reflect.String:
			if sortOpt.Value > 0 {
				return iVal.String() < jVal.String()
			}
			return iVal.String() > jVal.String()

		default:
			if iVal.IsZero() {
				return false
			}
			switch iVal.Type() {

			case reflect.ValueOf(time.Now()).Type():
				if sortOpt.Value > 0 {
					return iVal.Interface().(time.Time).Before(jVal.Interface().(time.Time))
				}
				return iVal.Interface().(time.Time).After(jVal.Interface().(time.Time))

			case reflect.ValueOf(timestamppb.Now()).Type():
				iStamp := iVal.Interface().(*timestamppb.Timestamp).AsTime()
				jStamp := jVal.Interface().(*timestamppb.Timestamp).AsTime()
				if sortOpt.Value > 0 {
					return iStamp.Before(jStamp)
				}
				return iStamp.After(jStamp)
			}
		}

		//log.Println("generateLessFunc() didn't match any cases")
		//log.Println(iVal)
		return false
	}
}

func matchFieldFunc(name string) func(string) bool {
	return func(to string) bool {
		return isLowerEqual(removeUnderscore(name), removeUnderscore(to))
	}
}

func (d *DB) Update(collection string, object interface{}, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}

	dataSlice := d.collectionMap[collection]
	for i, data := range *dataSlice {
		if compareInterfaceToFilter(data, filter) {
			return setValue(&(*dataSlice)[i], object)
		}
	}

	return errors.New("mock.DB.Update() error: no documents found")
}

func (d *DB) Upsert(collection string, object interface{}, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}
	dataSlice := d.collectionMap[collection]
	// if collection is empty just insert
	if dataSlice == nil {
		return d.Insert(collection, object)
	}
	for i, data := range *dataSlice {
		if compareInterfaceToFilter(data, filter) {
			return setValue(&(*dataSlice)[i], object)
		}
	}
	return d.Insert(collection, object)
}

func (d *DB) Delete(collection string, filter *db.Filter) error {
	if err := checkParams(collection, filter); err != nil {
		return fmt.Errorf("mock.DB.Update() error: %v", err)
	}
	dataSlice := d.collectionMap[collection]
	//dataSlicePtr := &dataSlice
	for i, data := range *dataSlice {
		if compareInterfaceToFilter(data, filter) {
			// get the slice value and slice value element
			sliceVal := reflect.ValueOf(d.collectionMap[collection])
			sliceValElem := sliceVal.Elem()

			// gather the slice around the object
			part1 := sliceValElem.Slice(0, i)
			part2 := sliceValElem.Slice(i+1, len(*dataSlice))

			// create a new slice and append
			sliceValElem.Set(reflect.AppendSlice(part1, part2))

			return nil
		}
	}
	return errors.New("mock.DB.Delete(): no documents found")
}

func (d *DB) Search(collection string, search string, fields []string, slice interface{}) error {
	dataSlice := d.collectionMap[collection]
	pointerVal := reflect.ValueOf(slice)
	sliceVal := pointerVal.Elem()
	for _, data := range *dataSlice {
		dataVal := reflect.ValueOf(data)
		for _, field := range fields {
			fieldValue := dataVal.FieldByNameFunc(matchFieldFunc(field))
			if containsLower(fieldValue.String(), search) {
				appendSliceVal(&sliceVal, data)
			}
		}
	}

	pointerVal.Elem().Set(sliceVal)
	return nil
}

func appendSliceVal(sliceVal *reflect.Value, data interface{}) {
	dataVal := reflect.ValueOf(data)
	// if the slice contains pointers to object
	if reflect.TypeOf(sliceVal.Interface()).Elem().Kind() == reflect.Ptr {
		p := reflect.New(reflect.TypeOf(data))
		p.Elem().Set(reflect.ValueOf(data))

		dataVal = p
	}

	*sliceVal = reflect.Append(*sliceVal, dataVal)
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

	if aVal.Kind() == reflect.Ptr {
		aVal = aVal.Elem()
	}

	for filterKey, filterVal := range *filter {
		fieldVal := aVal.FieldByNameFunc(matchFieldFunc(filterKey))
		//if fieldVal.IsZero() {
		//	continue
		//}
		if !isEqual(reflect.ValueOf(filterVal), fieldVal) {
			//log.Printf("compareInterfaceToFilter() not equal: \n\t%v\n\t%v\n", filterVal, fieldVal.Interface())
			return false
		}
	}
	return true
}

func isEqual(a, b reflect.Value) bool {
	if a.Type() == b.Type() {
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
	if isIntKind(a) && isIntKind(b) {
		return a.Int() == b.Int()
	}

	//log.Printf("isEqual():incorrect types: %v ?= %v\n", a.Type(), b.Type())
	return a.Interface() == b.Interface()
}

func isIntKind(v reflect.Value) bool {
	return v.Kind() == reflect.Int || v.Kind() == reflect.Int8 ||
		v.Kind() == reflect.Int16 || v.Kind() == reflect.Int32 ||
		v.Kind() == reflect.Int64
}

func containsLower(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func isLowerEqual(a, b string) bool {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b)) == 0
}

func removeUnderscore(s string) string {
	return strings.ReplaceAll(s, "_", "")
}

func setValue(into, datafrom interface{}) error {
	if reflect.TypeOf(into).Kind() != reflect.Ptr {
		return errors.New("input object is not type pointer")
	}
	datafromVal := reflect.ValueOf(datafrom)
	if datafromVal.Kind() == reflect.Ptr {
		datafromVal = datafromVal.Elem()
	}
	reflect.ValueOf(into).Elem().Set(datafromVal)
	return nil
}
