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

import "testing"

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Title string
}

func TestTableGetStruct(t *testing.T) {
	db := &DB{}
	if err := db.BindModel("users", User{}, "id"); err != nil {
		t.Fatal(err)
	}
	b := &Batch{DB: db}
	b.PutStruct(&User{ID: 1, Name: "Peter"})
	b.PutStruct(User{ID: 2, Name: "Spencer", Title: "CEO"})
	b.GetStruct(&User{ID: 1}, "name")
	b.GetStruct(&User{ID: 2})
}
