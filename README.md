# stockpile

** Current state: ALPHA **

Interface to different databases, meant to allow easy mocking through the stockpile/mock package.

Useful in two ways:
1. Setup a persistance layer between data and database using the database interface
2. Use the mock database to run unit tests to test only specific business code


TODO:
- [x] Support limit option
- [x] Support skip option
- [/] Support sort option
	- [x] Int sort
	- [ ] String sort
	- [ ] Time sort

- [x] Open(ctx context.Context) error
- [x] Close(ctx context.Context) error

- [x] Insert(collection string, object interface{}) error
- [ ] InsertMany(collection string, slice interface{}) error
- [x] FindOne(collection string, object interface{}, filter *Filter, opts *Options) error
- [x] FindAll(collection string, object interface{}, filter *Filter, opts *Options) error
- [x] Update(collection string, object interface{}, filter *Filter) error
- [x] Upsert(collection string, object interface{}, filter *Filter) error
- [x] Delete(collection string, filter *Filter) error
- [x] Search(collection, search string, fields []string, object interface{}) error

# filter matching will remove underscores in field names
