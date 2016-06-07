/*
   Copyright 2016 GitHub Inc.
	 See https://github.com/github/gh-ost/blob/master/LICENSE
*/

package sql

import (
	"testing"

	"reflect"
	"regexp"
	"strings"

	"github.com/outbrain/golib/log"
	test "github.com/outbrain/golib/tests"
)

var (
	spacesRegexp = regexp.MustCompile(`[ \t\n\r]+`)
)

func init() {
	log.SetLevel(log.ERROR)
}

func normalizeQuery(name string) string {
	name = strings.Replace(name, "`", "", -1)
	name = spacesRegexp.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	return name
}

func TestEscapeName(t *testing.T) {
	names := []string{"my_table", `"my_table"`, "`my_table`"}
	for _, name := range names {
		escaped := EscapeName(name)
		test.S(t).ExpectEquals(escaped, "`my_table`")
	}
}

func TestBuildEqualsComparison(t *testing.T) {
	{
		columns := []string{"c1"}
		values := []string{"@v1"}
		comparison, err := BuildEqualsComparison(columns, values)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` = @v1))")
	}
	{
		columns := []string{"c1", "c2"}
		values := []string{"@v1", "@v2"}
		comparison, err := BuildEqualsComparison(columns, values)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` = @v1) and (`c2` = @v2))")
	}
	{
		columns := []string{"c1"}
		values := []string{"@v1", "@v2"}
		_, err := BuildEqualsComparison(columns, values)
		test.S(t).ExpectNotNil(err)
	}
	{
		columns := []string{}
		values := []string{}
		_, err := BuildEqualsComparison(columns, values)
		test.S(t).ExpectNotNil(err)
	}
}

func TestBuildEqualsPreparedComparison(t *testing.T) {
	{
		columns := []string{"c1", "c2"}
		comparison, err := BuildEqualsPreparedComparison(columns)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` = ?) and (`c2` = ?))")
	}
}

func TestBuildSetPreparedClause(t *testing.T) {
	{
		columns := []string{"c1"}
		clause, err := BuildSetPreparedClause(columns)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(clause, "`c1`=?")
	}
	{
		columns := []string{"c1", "c2"}
		clause, err := BuildSetPreparedClause(columns)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(clause, "`c1`=?, `c2`=?")
	}
	{
		columns := []string{}
		_, err := BuildSetPreparedClause(columns)
		test.S(t).ExpectNotNil(err)
	}
}

func TestBuildRangeComparison(t *testing.T) {
	{
		columns := []string{"c1"}
		values := []string{"@v1"}
		args := []interface{}{3}
		comparison, explodedArgs, err := BuildRangeComparison(columns, values, args, LessThanComparisonSign)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` < @v1))")
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3}))
	}
	{
		columns := []string{"c1"}
		values := []string{"@v1"}
		args := []interface{}{3}
		comparison, explodedArgs, err := BuildRangeComparison(columns, values, args, LessThanOrEqualsComparisonSign)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` < @v1) or ((`c1` = @v1)))")
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3}))
	}
	{
		columns := []string{"c1", "c2"}
		values := []string{"@v1", "@v2"}
		args := []interface{}{3, 17}
		comparison, explodedArgs, err := BuildRangeComparison(columns, values, args, LessThanComparisonSign)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` < @v1) or (((`c1` = @v1)) AND (`c2` < @v2)))")
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17}))
	}
	{
		columns := []string{"c1", "c2"}
		values := []string{"@v1", "@v2"}
		args := []interface{}{3, 17}
		comparison, explodedArgs, err := BuildRangeComparison(columns, values, args, LessThanOrEqualsComparisonSign)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` < @v1) or (((`c1` = @v1)) AND (`c2` < @v2)) or ((`c1` = @v1) and (`c2` = @v2)))")
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17, 3, 17}))
	}
	{
		columns := []string{"c1", "c2", "c3"}
		values := []string{"@v1", "@v2", "@v3"}
		args := []interface{}{3, 17, 22}
		comparison, explodedArgs, err := BuildRangeComparison(columns, values, args, LessThanOrEqualsComparisonSign)
		test.S(t).ExpectNil(err)
		test.S(t).ExpectEquals(comparison, "((`c1` < @v1) or (((`c1` = @v1)) AND (`c2` < @v2)) or (((`c1` = @v1) and (`c2` = @v2)) AND (`c3` < @v3)) or ((`c1` = @v1) and (`c2` = @v2) and (`c3` = @v3)))")
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17, 3, 17, 22, 3, 17, 22}))
	}
	{
		columns := []string{"c1"}
		values := []string{"@v1", "@v2"}
		args := []interface{}{3, 17}
		_, _, err := BuildRangeComparison(columns, values, args, LessThanOrEqualsComparisonSign)
		test.S(t).ExpectNotNil(err)
	}
	{
		columns := []string{}
		values := []string{}
		args := []interface{}{}
		_, _, err := BuildRangeComparison(columns, values, args, LessThanOrEqualsComparisonSign)
		test.S(t).ExpectNotNil(err)
	}
}

func TestBuildRangeInsertQuery(t *testing.T) {
	databaseName := "mydb"
	originalTableName := "tbl"
	ghostTableName := "ghost"
	sharedColumns := []string{"id", "name", "position"}
	{
		uniqueKey := "PRIMARY"
		uniqueKeyColumns := []string{"id"}
		rangeStartValues := []string{"@v1s"}
		rangeEndValues := []string{"@v1e"}
		rangeStartArgs := []interface{}{3}
		rangeEndArgs := []interface{}{103}

		query, explodedArgs, err := BuildRangeInsertQuery(databaseName, originalTableName, ghostTableName, sharedColumns, uniqueKey, uniqueKeyColumns, rangeStartValues, rangeEndValues, rangeStartArgs, rangeEndArgs, true, false)
		test.S(t).ExpectNil(err)
		expected := `
				insert /* gh-ost mydb.tbl */ ignore into mydb.ghost (id, name, position)
				(select id, name, position from mydb.tbl force index (PRIMARY)
					where (((id > @v1s) or ((id = @v1s))) and ((id < @v1e) or ((id = @v1e))))
				)
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 103, 103}))
	}
	{
		uniqueKey := "name_position_uidx"
		uniqueKeyColumns := []string{"name", "position"}
		rangeStartValues := []string{"@v1s", "@v2s"}
		rangeEndValues := []string{"@v1e", "@v2e"}
		rangeStartArgs := []interface{}{3, 17}
		rangeEndArgs := []interface{}{103, 117}

		query, explodedArgs, err := BuildRangeInsertQuery(databaseName, originalTableName, ghostTableName, sharedColumns, uniqueKey, uniqueKeyColumns, rangeStartValues, rangeEndValues, rangeStartArgs, rangeEndArgs, true, false)
		test.S(t).ExpectNil(err)
		expected := `
				insert /* gh-ost mydb.tbl */ ignore into mydb.ghost (id, name, position)
				(select id, name, position from mydb.tbl force index (name_position_uidx)
				  where (((name > @v1s) or (((name = @v1s)) AND (position > @v2s)) or ((name = @v1s) and (position = @v2s))) and ((name < @v1e) or (((name = @v1e)) AND (position < @v2e)) or ((name = @v1e) and (position = @v2e))))
				)
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17, 3, 17, 103, 103, 117, 103, 117}))
	}
}

func TestBuildRangeInsertPreparedQuery(t *testing.T) {
	databaseName := "mydb"
	originalTableName := "tbl"
	ghostTableName := "ghost"
	sharedColumns := []string{"id", "name", "position"}
	{
		uniqueKey := "name_position_uidx"
		uniqueKeyColumns := []string{"name", "position"}
		rangeStartArgs := []interface{}{3, 17}
		rangeEndArgs := []interface{}{103, 117}

		query, explodedArgs, err := BuildRangeInsertPreparedQuery(databaseName, originalTableName, ghostTableName, sharedColumns, uniqueKey, uniqueKeyColumns, rangeStartArgs, rangeEndArgs, true, true)
		test.S(t).ExpectNil(err)
		expected := `
				insert /* gh-ost mydb.tbl */ ignore into mydb.ghost (id, name, position)
				(select id, name, position from mydb.tbl force index (name_position_uidx)
				  where (((name > ?) or (((name = ?)) AND (position > ?)) or ((name = ?) and (position = ?))) and ((name < ?) or (((name = ?)) AND (position < ?)) or ((name = ?) and (position = ?))))
				lock in share mode )
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17, 3, 17, 103, 103, 117, 103, 117}))
	}
}

func TestBuildUniqueKeyRangeEndPreparedQuery(t *testing.T) {
	databaseName := "mydb"
	originalTableName := "tbl"
	var chunkSize int64 = 500
	{
		uniqueKeyColumns := []string{"name", "position"}
		rangeStartArgs := []interface{}{3, 17}
		rangeEndArgs := []interface{}{103, 117}

		query, explodedArgs, err := BuildUniqueKeyRangeEndPreparedQuery(databaseName, originalTableName, uniqueKeyColumns, rangeStartArgs, rangeEndArgs, chunkSize, false, "test")
		test.S(t).ExpectNil(err)
		expected := `
				select /* gh-ost mydb.tbl test */ name, position
				  from (
				    select
				        name, position
				      from
				        mydb.tbl
				      where ((name > ?) or (((name = ?)) AND (position > ?))) and ((name < ?) or (((name = ?)) AND (position < ?)) or ((name = ?) and (position = ?)))
				      order by
				        name asc, position asc
				      limit 500
				  ) select_osc_chunk
				order by
				  name desc, position desc
				limit 1
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(explodedArgs, []interface{}{3, 3, 17, 103, 103, 117, 103, 117}))
	}
}

func TestBuildUniqueKeyMinValuesPreparedQuery(t *testing.T) {
	databaseName := "mydb"
	originalTableName := "tbl"
	uniqueKeyColumns := []string{"name", "position"}
	{
		query, err := BuildUniqueKeyMinValuesPreparedQuery(databaseName, originalTableName, uniqueKeyColumns)
		test.S(t).ExpectNil(err)
		expected := `
			select /* gh-ost mydb.tbl */ name, position
			  from
			    mydb.tbl
			  order by
			    name asc, position asc
			  limit 1
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
	}
	{
		query, err := BuildUniqueKeyMaxValuesPreparedQuery(databaseName, originalTableName, uniqueKeyColumns)
		test.S(t).ExpectNil(err)
		expected := `
			select /* gh-ost mydb.tbl */ name, position
			  from
			    mydb.tbl
			  order by
			    name desc, position desc
			  limit 1
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
	}
}

func TestBuildDMLDeleteQuery(t *testing.T) {
	databaseName := "mydb"
	tableName := "tbl"
	tableColumns := NewColumnList([]string{"id", "name", "rank", "position", "age"})
	args := []interface{}{3, "testname", "first", 17, 23}
	{
		uniqueKeyColumns := NewColumnList([]string{"position"})

		query, uniqueKeyArgs, err := BuildDMLDeleteQuery(databaseName, tableName, tableColumns, uniqueKeyColumns, args)
		test.S(t).ExpectNil(err)
		expected := `
			delete /* gh-ost mydb.tbl */
				from
					mydb.tbl
				where
					((position = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{17}))
	}
	{
		uniqueKeyColumns := NewColumnList([]string{"name", "position"})

		query, uniqueKeyArgs, err := BuildDMLDeleteQuery(databaseName, tableName, tableColumns, uniqueKeyColumns, args)
		test.S(t).ExpectNil(err)
		expected := `
			delete /* gh-ost mydb.tbl */
				from
					mydb.tbl
				where
					((name = ?) and (position = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{"testname", 17}))
	}
	{
		uniqueKeyColumns := NewColumnList([]string{"position", "name"})

		query, uniqueKeyArgs, err := BuildDMLDeleteQuery(databaseName, tableName, tableColumns, uniqueKeyColumns, args)
		test.S(t).ExpectNil(err)
		expected := `
			delete /* gh-ost mydb.tbl */
				from
					mydb.tbl
				where
					((position = ?) and (name = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{17, "testname"}))
	}
	{
		uniqueKeyColumns := NewColumnList([]string{"position", "name"})
		args := []interface{}{"first", 17}

		_, _, err := BuildDMLDeleteQuery(databaseName, tableName, tableColumns, uniqueKeyColumns, args)
		test.S(t).ExpectNotNil(err)
	}
}

func TestBuildDMLInsertQuery(t *testing.T) {
	databaseName := "mydb"
	tableName := "tbl"
	tableColumns := NewColumnList([]string{"id", "name", "rank", "position", "age"})
	args := []interface{}{3, "testname", "first", 17, 23}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		query, sharedArgs, err := BuildDMLInsertQuery(databaseName, tableName, tableColumns, sharedColumns, args)
		test.S(t).ExpectNil(err)
		expected := `
			replace /* gh-ost mydb.tbl */
				into mydb.tbl
					(id, name, position, age)
				values
					(?, ?, ?, ?)
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{3, "testname", 17, 23}))
	}
	{
		sharedColumns := NewColumnList([]string{"position", "name", "age", "id"})
		query, sharedArgs, err := BuildDMLInsertQuery(databaseName, tableName, tableColumns, sharedColumns, args)
		test.S(t).ExpectNil(err)
		expected := `
			replace /* gh-ost mydb.tbl */
				into mydb.tbl
					(position, name, age, id)
				values
					(?, ?, ?, ?)
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{17, "testname", 23, 3}))
	}
	{
		sharedColumns := NewColumnList([]string{"position", "name", "surprise", "id"})
		_, _, err := BuildDMLInsertQuery(databaseName, tableName, tableColumns, sharedColumns, args)
		test.S(t).ExpectNotNil(err)
	}
	{
		sharedColumns := NewColumnList([]string{})
		_, _, err := BuildDMLInsertQuery(databaseName, tableName, tableColumns, sharedColumns, args)
		test.S(t).ExpectNotNil(err)
	}
}

func TestBuildDMLUpdateQuery(t *testing.T) {
	databaseName := "mydb"
	tableName := "tbl"
	tableColumns := NewColumnList([]string{"id", "name", "rank", "position", "age"})
	valueArgs := []interface{}{3, "testname", "newval", 17, 23}
	whereArgs := []interface{}{3, "testname", "findme", 17, 56}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{"position"})
		query, sharedArgs, uniqueKeyArgs, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNil(err)
		expected := `
			update /* gh-ost mydb.tbl */
			  mydb.tbl
					set id=?, name=?, position=?, age=?
				where
					((position = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{3, "testname", 17, 23}))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{17}))
	}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{"position", "name"})
		query, sharedArgs, uniqueKeyArgs, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNil(err)
		expected := `
			update /* gh-ost mydb.tbl */
			  mydb.tbl
					set id=?, name=?, position=?, age=?
				where
					((position = ?) and (name = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{3, "testname", 17, 23}))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{17, "testname"}))
	}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{"age"})
		query, sharedArgs, uniqueKeyArgs, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNil(err)
		expected := `
			update /* gh-ost mydb.tbl */
			  mydb.tbl
					set id=?, name=?, position=?, age=?
				where
					((age = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{3, "testname", 17, 23}))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{56}))
	}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{"age", "position", "id", "name"})
		query, sharedArgs, uniqueKeyArgs, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNil(err)
		expected := `
			update /* gh-ost mydb.tbl */
			  mydb.tbl
					set id=?, name=?, position=?, age=?
				where
					((age = ?) and (position = ?) and (id = ?) and (name = ?))
		`
		test.S(t).ExpectEquals(normalizeQuery(query), normalizeQuery(expected))
		test.S(t).ExpectTrue(reflect.DeepEqual(sharedArgs, []interface{}{3, "testname", 17, 23}))
		test.S(t).ExpectTrue(reflect.DeepEqual(uniqueKeyArgs, []interface{}{56, 17, 3, "testname"}))
	}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{"age", "surprise"})
		_, _, _, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNotNil(err)
	}
	{
		sharedColumns := NewColumnList([]string{"id", "name", "position", "age"})
		uniqueKeyColumns := NewColumnList([]string{})
		_, _, _, err := BuildDMLUpdateQuery(databaseName, tableName, tableColumns, sharedColumns, uniqueKeyColumns, valueArgs, whereArgs)
		test.S(t).ExpectNotNil(err)
	}
}