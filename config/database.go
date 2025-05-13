package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"todorist/env"
	"todorist/pkg/exception"

	_ "github.com/lib/pq"
)

type DB struct {
	ctx    context.Context
	db     *sql.DB
	UserId string
	tx     *sql.Tx
}

func NewDB(ctx context.Context, dbConfig string) *DB {
	return &DB{
		ctx: ctx,
		db:  dbConnect(dbConfig),
	}
}

func dbConnect(connectionString string) *sql.DB {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	log.Printf("Database connected! Host: %s Port: %d DB: %s", env.PgHost, env.PgPort, env.PgDatabase)
	return db
}

func (db *DB) Close() {
	err := db.db.Close()
	if err != nil {
		log.Fatal("got error when closing the DB connection", err)
	}
}

func (db *DB) SetUserId(userId string) {
	db.UserId = userId
}

func (db *DB) GetUserId() string {
	return db.UserId
}

func (db *DB) replaceQuery(query string, params map[string]any, startFrom ...uint) (string, []any) {
	retparams := make([]any, 0)
	if len(params) == 0 {
		return query, retparams
	}
	var iterator uint = 1
	if len(startFrom) > 0 {
		iterator = startFrom[0]
	}
	pattern := `\$\<([^>:]+):?([^>:]+)?\>`
	m := regexp.MustCompile(pattern)
	replacedByte := m.ReplaceAllFunc([]byte(query), func(s []byte) []byte {
		strfind := m.FindAllSubmatch(s, -1)
		paramName := string(strfind[0][1])
		paramType := string(strfind[0][2])

		val, ok := params[paramName]
		if !ok {
			log.Panicf("parameter %s not found", paramName)
			return s
		}

		switch paramType {
		case "list":
			templates := make([]string, 0)
			arrVal := reflect.ValueOf(val)
			if arrVal.Kind() == reflect.Slice {
				for i := 0; i < arrVal.Len(); i++ {
					retparams = append(retparams, arrVal.Index(i).Interface())
					templates = append(templates, fmt.Sprint("$", iterator))
					iterator++
				}
			}
			return []byte(strings.Join(templates, ", "))
		case "raw":
			return []byte(fmt.Sprintf("%v", val))
		default:
			retparams = append(retparams, val)
			stri := fmt.Sprintf("$%d", iterator)
			iterator++
			return []byte(stri)
		}
	})

	return string(replacedByte), retparams
}

type SelectOption struct {
	checkNotFound *string
}

func (db *DB) SelectOne(query string, result interface{}, args map[string]any, options ...func(*SelectOption)) error {
	opt := &SelectOption{}
	for _, option := range options {
		option(opt)
	}

	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr {
		return errors.New("result must be pointer")
	}
	resultVal = resultVal.Elem()
	if resultVal.Kind() != reflect.Struct {
		return errors.New("result must be struct")
	}

	repquery, repargs := db.replaceQuery(query, args)
	log.Println(repquery)
	log.Println(repargs...)
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(repquery, repargs...)
	} else {
		rows, err = db.db.Query(repquery, repargs...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	var rowCount uint = 0

	structAddr := make(map[string]any, 0)
	for i := 0; i < resultVal.NumField(); i++ {
		tag, ok := resultVal.Type().Field(i).Tag.Lookup("db")
		if !ok {
			continue
		}
		addr := resultVal.Field(i).Addr().Interface()
		structAddr[tag] = addr
	}

	for rows.Next() {
		rowCount++
		if rowCount > 1 {
			return errors.New("too many rows returned")
		}
		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := structAddr[columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
	}

	if opt.checkNotFound != nil && rowCount == 0 {
		return &exception.NotFoundException{
			Message: *opt.checkNotFound,
		}
	}

	return nil
}

func (db *DB) SelectMany(query string, result interface{}, args map[string]any) error {
	response := reflect.ValueOf(result)
	if response.Kind() != reflect.Ptr {
		return errors.New("result must be pointer")
	}
	resultVal := response.Elem()
	if resultVal.Kind() != reflect.Slice {
		return errors.New("result must be slice")
	}
	elementType := resultVal.Type().Elem()
	if elementType.Kind() != reflect.Struct {
		return fmt.Errorf("result must be slice of struct, got %s", elementType.Kind().String())
	}

	repquery, repargs := db.replaceQuery(query, args)
	log.Println(repquery)
	log.Println(repargs...)
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(repquery, repargs...)
	} else {
		rows, err = db.db.Query(repquery, repargs...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	for rows.Next() {
		structAddr := make(map[string]any, 0)
		structValue := reflect.New(elementType).Elem()
		for i := 0; i < elementType.NumField(); i++ {
			tag, ok := elementType.Field(i).Tag.Lookup("db")
			if !ok {
				continue
			}
			addr := structValue.Field(i).Addr().Interface()
			structAddr[tag] = addr
		}

		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := structAddr[columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
		resultVal = reflect.Append(resultVal, structValue)
	}

	response.Elem().Set(resultVal)

	return nil
}

type MutateOption struct {
	IsContainUserId bool
	ConflictKey     string
}

func WithoutUserId() func(*MutateOption) {
	return func(option *MutateOption) {
		option.IsContainUserId = false
	}
}

func (db *DB) InsertOne(data interface{}, tableName string, returning interface{}, options ...func(*MutateOption)) error {
	mo := MutateOption{IsContainUserId: true}
	for _, option := range options {
		option(&mo)
	}
	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}
	if dataVal.Kind() != reflect.Struct {
		return errors.New("data must be struct")
	}

	columns := make([]string, 0)
	templates := make([]string, 0)
	values := make([]any, 0)
	updateSet := make([]string, 0)

	tmpltCnt := 1
	for i := 0; i < dataVal.NumField(); i++ {
		tagValue, ok := dataVal.Type().Field(i).Tag.Lookup("db")
		if !ok {
			continue
		}

		arrayTag := strings.Split(tagValue, ",")
		value := dataVal.Field(i).Interface()

		isEmpty := false
		if dataVal.Field(i).Kind() == reflect.Ptr {
			isEmpty = dataVal.Field(i).IsNil()
		} else {
			isEmpty = reflect.DeepEqual(value, reflect.Zero(dataVal.Field(i).Type()).Interface())
		}

		if slices.Contains(arrayTag, "omitempty") && isEmpty {
			continue
		} else if slices.Contains(arrayTag, "nullable") && isEmpty {
			columns = append(columns, arrayTag[0])
			templates = append(templates, "NULL")
			continue
		}

		columns = append(columns, arrayTag[0])
		templates = append(templates, fmt.Sprintf("$%d", tmpltCnt))
		values = append(values, value)
		tmpltCnt++

		updateSet = append(updateSet, fmt.Sprintf("%s = EXCLUDED.%s", arrayTag[0], arrayTag[0]))
	}

	if mo.IsContainUserId {
		strUserId := fmt.Sprintf("'%s'", db.UserId)
		columns = append(columns, "created_by", "updated_by")
		templates = append(templates, strUserId, strUserId)
	}

	returnKey := make([]string, 0)
	returnAddr := make(map[string]any, 0)

	if returning != nil {
		returningVal := reflect.ValueOf(returning)
		if returningVal.Kind() != reflect.Ptr {
			return errors.New("returning must be pointer")
		}
		returningVal = returningVal.Elem()
		if returningVal.Kind() != reflect.Struct {
			return errors.New("returning must be struct")
		}
		for i := 0; i < returningVal.NumField(); i++ {
			tag, ok := returningVal.Type().Field(i).Tag.Lookup("db")
			if !ok {
				continue
			}
			returnKey = append(returnKey, tag)
			returnAddr[tag] = returningVal.Field(i).Addr().Interface()
		}
	}

	returningStr := ""
	if len(returnKey) > 0 {
		returningStr = "\nRETURNING " + strings.Join(returnKey, ", ")
	}

	upsertStr := ""
	if mo.ConflictKey != "" {
		upsertStr += fmt.Sprintf("\nON CONFLICT (%s) DO UPDATE SET %s", mo.ConflictKey, strings.Join(updateSet, ", "))
	}

	query := fmt.Sprintf("INSERT INTO %s(%s)\nVALUES(%s)", tableName, strings.Join(columns, ", "), strings.Join(templates, ", "))
	query += upsertStr
	query += returningStr

	log.Println(query)
	log.Println(values...)
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(query, values...)
	} else {
		rows, err = db.db.Query(query, values...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	if len(returnKey) == 0 {
		return nil
	}

	for rows.Next() {
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := returnAddr[columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) InsertMany(data interface{}, tableName string, returning interface{}, options ...func(*MutateOption)) error {
	mo := MutateOption{IsContainUserId: true}
	for _, option := range options {
		option(&mo)
	}
	dataVal := reflect.ValueOf(data)
	var dataSliceVal reflect.Value
	if dataVal.Kind() == reflect.Ptr {
		dataSliceVal = dataVal.Elem()
	} else {
		dataSliceVal = dataVal
	}
	if dataSliceVal.Kind() != reflect.Slice {
		return errors.New("data must be slice")
	}
	elementType := dataSliceVal.Type().Elem()
	if elementType.Kind() != reflect.Struct {
		return fmt.Errorf("data must be slice of struct, got %s", elementType.Kind().String())
	}

	columns := make([]string, 0)
	templates := make([]string, 0)
	values := make([]any, 0)

	tmpltCnt := 1
	for i := 0; i < dataSliceVal.Len(); i++ {
		for j := 0; j < elementType.NumField(); j++ {
			tag, ok := elementType.Field(j).Tag.Lookup("db")
			if !ok {
				continue
			}
			if i == 0 {
				columns = append(columns, fmt.Sprintf("\"%s\"", tag))
			}
			arrayTag := strings.Split(tag, ",")
			value := dataSliceVal.Index(i).Field(j).Interface()

			isEmpty := false
			element := dataSliceVal.Index(i)
			if element.Kind() == reflect.Ptr {
				isEmpty = element.IsNil()
			} else {
				isEmpty = reflect.DeepEqual(value, reflect.Zero(element.Type()).Interface())
			}

			if slices.Contains(arrayTag, "omitempty") && isEmpty {
				continue
			} else if slices.Contains(arrayTag, "nullable") && isEmpty {
				if i == 0 {
					columns = append(columns, arrayTag[0])
				}
				templates = append(templates, "NULL")
				continue
			}
			templates = append(templates, fmt.Sprintf("$%d", tmpltCnt))
			values = append(values, value)
			tmpltCnt++
		}
		if mo.IsContainUserId {
			strUserId := fmt.Sprintf("'%s'", db.UserId)
			templates = append(templates, strUserId, strUserId)
		}
	}
	if mo.IsContainUserId {
		columns = append(columns, "created_by", "updated_by")
	}

	returnKey := make([]string, 0)
	returnAddr := make([]map[string]any, 0)

	var returningVal reflect.Value
	var returningSliceVal reflect.Value
	var returningElType reflect.Type

	if returning != nil {
		returningVal = reflect.ValueOf(returning)
		if returningVal.Kind() != reflect.Ptr {
			return errors.New("returning must be pointer")
		}
		returningSliceVal = returningVal.Elem()
		if returningSliceVal.Kind() != reflect.Struct {
			return errors.New("returning must be struct")
		}
		returningElType = returningSliceVal.Type().Elem()
		if returningElType.Kind() != reflect.Struct {
			return fmt.Errorf("result must be slice of struct, got %s", returningElType.Kind().String())
		}
		for i := 0; i < returningSliceVal.Len(); i++ {
			tempAddr := make(map[string]any, 0)
			structClone := reflect.New(returningElType).Elem()
			for j := 0; j < returningElType.NumField(); j++ {
				tag, ok := returningElType.Field(j).Tag.Lookup("db")
				if !ok {
					continue
				}
				if i == 0 {
					returnKey = append(returnKey, tag)
				}
				tempAddr[tag] = structClone.Field(j).Addr().Interface()
			}
			returningSliceVal = reflect.Append(returningSliceVal, structClone)
			returnAddr = append(returnAddr, tempAddr)
		}
	}

	returningStr := ""
	if len(returnKey) > 0 {
		returningStr = "\nRETURNING " + strings.Join(returnKey, ", ")
	}
	getValuesTemplate := func(t []string, c int) string {
		r := make([]string, 0)
		for i := 0; i < dataSliceVal.Len(); i++ {
			tmp := make([]string, 0)
			for j := 0; j < c; j++ {
				tmp = append(tmp, t[i*c+j])
			}
			r = append(r, "("+strings.Join(tmp, ", ")+")")
		}
		return strings.Join(r, ",\n		")
	}
	query := fmt.Sprintf("INSERT INTO %s(%s)\nVALUES%s%s", tableName, strings.Join(columns, ", "), getValuesTemplate(templates, len(columns)), returningStr)
	log.Println(query)
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(query, values...)
	} else {
		rows, err = db.db.Query(query, values...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	if returning == nil {
		return nil
	}
	irow := 0
	for rows.Next() {
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := returnAddr[irow][columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
		irow++
	}
	returningVal.Elem().Set(returningSliceVal)

	return nil
}

func (db *DB) Update(data interface{}, tableName string, where string, params map[string]any, returning interface{}, options ...func(*MutateOption)) error {
	mo := MutateOption{IsContainUserId: true}
	for _, option := range options {
		option(&mo)
	}
	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}
	if dataVal.Kind() != reflect.Struct {
		return errors.New("data must be struct")
	}

	columns := make([]string, 0)
	templates := make([]string, 0)
	values := make([]any, 0)

	placeholderIndex := 1
	for i := 0; i < dataVal.NumField(); i++ {
		tagValue, ok := dataVal.Type().Field(i).Tag.Lookup("db")
		if !ok {
			continue
		}
		arrayTag := strings.Split(tagValue, ",")
		value := dataVal.Field(i).Interface()

		isEmpty := false
		if dataVal.Field(i).Kind() == reflect.Ptr {
			isEmpty = dataVal.Field(i).IsNil()
		} else {
			isEmpty = reflect.DeepEqual(value, reflect.Zero(dataVal.Field(i).Type()).Interface())
		}

		if (slices.Contains(arrayTag, "omitempty") && isEmpty) || slices.Contains(arrayTag, "skip") {
			continue
		}

		columns = append(columns, arrayTag[0])
		if slices.Contains(arrayTag, "nullable") && isEmpty {
			templates = append(templates, "NULL")
			continue
		} else if slices.Contains(arrayTag, "raw") {
			templates = append(templates, value.(string))
			continue
		}

		templates = append(templates, fmt.Sprintf("$%d", placeholderIndex))
		values = append(values, value)
		placeholderIndex++
	}

	if mo.IsContainUserId {
		strUserId := fmt.Sprintf("'%s'", db.UserId)
		columns = append(columns, "updated_by")
		templates = append(templates, strUserId)
	}

	columns = append(columns, "updated_at")
	templates = append(templates, "now()")

	returnKey := make([]string, 0)
	returnAddr := make([]map[string]any, 0)

	// var returningVal reflect.Value
	// var returningSliceVal reflect.Value
	// var returningElType reflect.Type

	// if returning != nil {
	// 	returningVal = reflect.ValueOf(returning)
	// 	if returningVal.Kind() != reflect.Ptr {
	// 		return errors.New("returning must be pointer")
	// 	}
	// 	returningSliceVal = returningVal.Elem()
	// 	if returningSliceVal.Kind() != reflect.Struct {
	// 		return errors.New("returning must be struct")
	// 	}
	// 	returningElType = returningSliceVal.Type().Elem()
	// 	if returningElType.Kind() != reflect.Struct {
	// 		return fmt.Errorf("result must be slice of struct, got %s", returningElType.Kind().String())
	// 	}
	// 	for i := 0; i < returningSliceVal.Len(); i++ {
	// 		tempAddr := make(map[string]any, 0)
	// 		structClone := reflect.New(returningElType).Elem()
	// 		for j := 0; j < returningElType.NumField(); j++ {
	// 			tag, ok := returningElType.Field(j).Tag.Lookup("db")
	// 			if !ok {
	// 				continue
	// 			}
	// 			returnKey = append(returnKey, tag)
	// 			tempAddr[tag] = structClone.Field(j).Addr().Interface()
	// 		}
	// 		returningSliceVal = reflect.Append(returningSliceVal, structClone)
	// 		returnAddr = append(returnAddr, tempAddr)
	// 	}
	// }

	// Handle both struct and slice of struct cases
	var returningVal reflect.Value
	isSingleStruct := false

	if returning != nil {
		returningVal = reflect.ValueOf(returning)
		if returningVal.Kind() != reflect.Ptr {
			return errors.New("returning must be a pointer")
		}

		returningVal = returningVal.Elem()

		if returningVal.Kind() == reflect.Struct {
			isSingleStruct = true
			tempAddr := make(map[string]any)
			for i := 0; i < returningVal.NumField(); i++ {
				tag, ok := returningVal.Type().Field(i).Tag.Lookup("db")
				if !ok {
					continue
				}
				returnKey = append(returnKey, tag)
				tempAddr[tag] = returningVal.Field(i).Addr().Interface()
			}
			returnAddr = append(returnAddr, tempAddr)
		} else if returningVal.Kind() == reflect.Slice {
			returningElType := returningVal.Type().Elem()
			if returningVal.IsNil() {
				returningVal.Set(reflect.MakeSlice(returningVal.Type(), 0, 0))
			}

			for i := 0; i < returningVal.Len(); i++ {
				tempAddr := make(map[string]any)
				structClone := reflect.New(returningElType).Elem()
				for j := 0; j < returningElType.NumField(); j++ {
					tag, ok := returningElType.Field(j).Tag.Lookup("db")
					if !ok {
						continue
					}
					returnKey = append(returnKey, tag)
					tempAddr[tag] = structClone.Field(j).Addr().Interface()
				}
				returnAddr = append(returnAddr, tempAddr)
			}
		} else {
			return errors.New("returning must be a pointer to struct or slice of structs")
		}
	}

	returningStr := ""
	if len(returnKey) > 0 {
		returningStr = "\nRETURNING " + strings.Join(returnKey, ", ")
	}
	setValueTemplate := func(c []string, t []string) string {
		r := make([]string, 0)
		for i := 0; i < len(c); i++ {
			r = append(r, fmt.Sprintf("%s = %v", c[i], t[i]))
		}
		return strings.Join(r, ",\n		")
	}
	query := fmt.Sprintf("UPDATE %s\n SET \n		%s\nWHERE %s%s", tableName, setValueTemplate(columns, templates), where, returningStr)
	repquery, repvalue := db.replaceQuery(query, params, uint(placeholderIndex))
	query = repquery
	values = append(values, repvalue...)
	log.Println(query)
	log.Println(values...)
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(query, values...)
	} else {
		rows, err = db.db.Query(query, values...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	if returning == nil {
		return nil
	}

	irow := 0
	for rows.Next() {
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := returnAddr[irow][columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
		irow++

		// for returning slice
		if !isSingleStruct && returningVal.Kind() == reflect.Slice {
			newStruct := reflect.New(returningVal.Type().Elem()).Elem()
			for key, addr := range returnAddr[0] {
				field := newStruct.FieldByNameFunc(func(fieldName string) bool {
					field, _ := newStruct.Type().FieldByName(fieldName)
					return field.Tag.Get("db") == key
				})
				if field.IsValid() {
					field.Set(reflect.ValueOf(addr).Elem())
				}
			}
			log.Println("Appending struct to slice:", newStruct.Interface())
			returningVal.Set(reflect.Append(returningVal, newStruct))
		}
	}


	if isSingleStruct {
		tempAddr := returnAddr[0]
		for key, addr := range tempAddr {
			field := returningVal.FieldByNameFunc(func(fieldName string) bool {
				field, _ := returningVal.Type().FieldByName(fieldName)
				return field.Tag.Get("db") == key
			})
			if field.IsValid() {
				field.Set(reflect.ValueOf(addr).Elem())
			}
		}
	}

	return nil
}

func (db *DB) Tx(f func(tx *DB) error) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	newDb := DB{
		UserId: db.UserId,
		tx:     tx,
	}
	if err := f(&newDb); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) SoftDelete(tableName string, where string, params map[string]any, returning interface{}) error {
	returnKey := make([]string, 0)
	returnAddr := make([]map[string]any, 0)

	var returningVal reflect.Value
	var returningSliceVal reflect.Value
	var returningElType reflect.Type

	if returning != nil {
		returningVal = reflect.ValueOf(returning)
		if returningVal.Kind() != reflect.Ptr {
			return errors.New("returning must be pointer")
		}
		returningSliceVal = returningVal.Elem()
		if returningSliceVal.Kind() != reflect.Slice {
			return errors.New("returning must be slice")
		}
		returningElType = returningSliceVal.Type().Elem()
		if returningElType.Kind() != reflect.Struct {
			return fmt.Errorf("returning must be slice of struct, got %s", returningElType.Kind().String())
		}
		for i := 0; i < returningSliceVal.Len(); i++ {
			tempAddr := make(map[string]any, 0)
			structClone := reflect.New(returningElType).Elem()
			for j := 0; j < returningElType.NumField(); j++ {
				tag, ok := returningElType.Field(j).Tag.Lookup("db")
				if !ok {
					continue
				}
				returnKey = append(returnKey, tag)
				tempAddr[tag] = structClone.Field(j).Addr().Interface()
			}
			returningSliceVal = reflect.Append(returningSliceVal, structClone)
			returnAddr = append(returnAddr, tempAddr)
		}
	}

	returningStr := ""
	if len(returnKey) > 0 {
		returningStr = "\nRETURNING " + strings.Join(returnKey, ", ")
	}

	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE %s%s", tableName, where, returningStr)
	repquery, repvalue := db.replaceQuery(query, params, 1)
	log.Println(repquery)
	log.Println(repvalue...)
	values := repvalue

	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(repquery, values...)
	} else {
		rows, err = db.db.Query(repquery, values...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	if returning == nil {
		return nil
	}

	irow := 0
	for rows.Next() {
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scans := make([]any, len(columnTypes))
		for i, columnType := range columnTypes {
			addr, ok := returnAddr[irow][columnType.Name()]
			if !ok {
				return fmt.Errorf("column %s not found in struct", columnType.Name())
			}
			scans[i] = addr
		}
		if err := rows.Scan(scans...); err != nil {
			return err
		}
		irow++
	}

	returningVal.Elem().Set(returningSliceVal)

	return nil
}
