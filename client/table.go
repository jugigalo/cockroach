// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Peter Mattis (peter@cockroachlabs.com)

package client

import (
	"bytes"
	"fmt"
	"reflect"
)

// This file contains the experimental Cockroach table-based interface. The
// contents will eventually be dispersed into {batch,db,txn}.go, but are
// collected here during initial development.

// Model ...
type Model struct {
	Name         string
	fields       fieldMap
	primaryKey   []string
	otherColumns []string
}

// BindModel ...
func (db *DB) BindModel(name string, obj interface{}, primaryKey ...string) error {
	t := deref(reflect.TypeOf(obj))
	if db.models == nil {
		db.models = make(map[reflect.Type]*Model)
	}
	if _, ok := db.models[t]; ok {
		return fmt.Errorf("%s: model '%T' already defined", name, obj)
	}
	m := &Model{
		Name:       name,
		fields:     getDBFields(t),
		primaryKey: primaryKey,
	}
	isPrimaryKey := make(map[string]bool)
	for _, k := range primaryKey {
		isPrimaryKey[k] = true
	}
	for col := range m.fields {
		if !isPrimaryKey[col] {
			m.otherColumns = append(m.otherColumns, col)
		}
	}
	db.models[t] = m
	return nil
}

func (db *DB) getModel(t reflect.Type) (*Model, error) {
	if model, ok := db.models[t]; ok {
		return model, nil
	}
	return nil, fmt.Errorf("unable to find model for '%s'", t)
}

// GetStruct ...
func (db *DB) GetStruct(obj interface{}, columns ...string) error {
	b := db.NewBatch()
	b.GetStruct(obj, columns...)
	_, err := runOneResult(db, b)
	return err
}

// PutStruct ...
func (db *DB) PutStruct(obj interface{}, columns ...string) error {
	b := db.NewBatch()
	b.PutStruct(obj, columns...)
	_, err := runOneResult(db, b)
	return err
}

// IncStruct ...
func (db *DB) IncStruct(obj interface{}, value int64, column string) error {
	b := db.NewBatch()
	b.IncStruct(obj, value, column)
	_, err := runOneResult(db, b)
	return err
}

// ScanStruct ...
func (db *DB) ScanStruct(start, end interface{}, maxRows int64) error {
	b := db.NewBatch()
	b.ScanStruct(start, end, maxRows)
	_, err := runOneResult(db, b)
	return err
}

// DelStruct ...
func (db *DB) DelStruct(obj interface{}, columns ...string) error {
	b := db.NewBatch()
	b.DelStruct(obj, columns...)
	_, err := runOneResult(db, b)
	return err
}

// GetStruct ...
func (txn *Txn) GetStruct(obj interface{}, columns ...string) error {
	b := txn.NewBatch()
	b.GetStruct(obj, columns...)
	_, err := runOneResult(txn, b)
	return err
}

// PutStruct ...
func (txn *Txn) PutStruct(obj interface{}, columns ...string) error {
	b := txn.NewBatch()
	b.PutStruct(obj, columns...)
	_, err := runOneResult(txn, b)
	return err
}

// IncStruct ...
func (txn *Txn) IncStruct(obj interface{}, value int64, column string) error {
	b := txn.NewBatch()
	b.IncStruct(obj, value, column)
	_, err := runOneResult(txn, b)
	return err
}

// ScanStruct ...
func (txn *Txn) ScanStruct(start, end interface{}, maxRows int64) error {
	b := txn.NewBatch()
	b.ScanStruct(start, end, maxRows)
	_, err := runOneResult(txn, b)
	return err
}

// DelStruct ...
func (txn *Txn) DelStruct(obj interface{}, columns ...string) error {
	b := txn.NewBatch()
	b.DelStruct(obj, columns...)
	_, err := runOneResult(txn, b)
	return err
}

// GetStruct ...
func (b *Batch) GetStruct(obj interface{}, columns ...string) {
	// 1. Find model
	// 2. Generate primary key
	// 3. Get specified non-primary key columns
	// 4. Unmarshal result

	v := reflect.ValueOf(obj)
	objT := v.Type()
	if objT.Kind() != reflect.Ptr {
		b.initResult(0, 0, fmt.Errorf("obj must be a pointer: %T", obj))
		return
	}
	objT = objT.Elem()
	m, err := b.DB.getModel(objT)
	if err != nil {
		b.initResult(0, 0, err)
		return
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "/%s", m.Name)

	v = reflect.Indirect(v)
	for _, col := range m.primaryKey {
		f, ok := m.fields[col]
		if !ok {
			panic(fmt.Errorf("%s: unable to find field %s", m.Name, col))
		}
		fmt.Fprintf(&buf, "/%v", v.FieldByIndex(f.Index).Interface())
	}

	if len(columns) == 0 {
		columns = m.otherColumns
	}
	for _, col := range columns {
		fmt.Printf("Get %s/%s\n", buf.String(), col)
	}

	b.initResult(0, 0, nil)
}

// PutStruct ...
func (b *Batch) PutStruct(obj interface{}, columns ...string) {
	// 1. Find model
	// 2. Generate primary key
	// 3. Put specified non-primary key columns

	v := reflect.Indirect(reflect.ValueOf(obj))
	m, err := b.DB.getModel(v.Type())
	if err != nil {
		b.initResult(0, 0, err)
		return
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "/%s", m.Name)

	for _, col := range m.primaryKey {
		f, ok := m.fields[col]
		if !ok {
			panic(fmt.Errorf("%s: unable to find field %s", m.Name, col))
		}
		fmt.Fprintf(&buf, "/%v", v.FieldByIndex(f.Index).Interface())
	}

	if len(columns) == 0 {
		columns = m.otherColumns
	}
	for _, col := range columns {
		f, ok := m.fields[col]
		if !ok {
			panic(fmt.Errorf("%s: unable to find field %s", m.Name, col))
		}
		fmt.Printf("Put %s/%s -> \"%s\"\n", buf.String(), col, v.FieldByIndex(f.Index))
	}

	b.initResult(0, 0, nil)
}

// IncStruct ...
func (b *Batch) IncStruct(obj interface{}, value int64, column string) {
	// 1. Find model
	// 2. Generate primary key
	// 3. Inc primary-key/column by value
}

// ScanStruct ...
func (b *Batch) ScanStruct(start, end interface{}, maxRows int64) {
	// 1. Find model
	// 2. Generate primary key for start and end
	// 3. Perform scan
	// 4. Unmarshal result
}

// DelStruct ...
func (b *Batch) DelStruct(obj interface{}, columns ...string) {
	// 1. Find model
	// 2. Generate primary key for specified columns
	// 3. Delete keys
}
