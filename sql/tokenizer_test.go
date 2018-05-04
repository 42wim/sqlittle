package sql

import (
	"errors"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/davecgh/go-spew/spew"
)

func TestTokens(t *testing.T) {
	testOK := func(sql string, want []token) {
		t.Helper()
		ts, err := tokenize(sql)
		if err != nil {
			t.Error(err)
			return
		}
		if have := ts; !reflect.DeepEqual(have, want) {
			t.Errorf("diff:\n%s", diff.LineDiff(spew.Sdump(want), spew.Sdump(have)))
		}
	}
	testError := func(sql string, want error) {
		t.Helper()
		_, err := tokenize(sql)
		if have := err; !reflect.DeepEqual(have, want) {
			t.Errorf("have %#v, want %#v", have, want)
		}
	}

	testOK(
		"foo foo_bar FoObAr foo1 _foo café",
		[]token{
			stoken(tBare, "foo"),
			stoken(tBare, "foo_bar"),
			stoken(tBare, "FoObAr"),
			stoken(tBare, "foo1"),
			stoken(tBare, "_foo"),
			stoken(tBare, "café"),
		},
	)

	testOK(
		"1 -12 +34",
		[]token{
			ntoken(tSignedNumber, 1),
			ntoken(tSignedNumber, -12),
			ntoken(tSignedNumber, +34),
		},
	)

	testOK(
		"create table foo",
		[]token{
			stoken(CREATE, "create"),
			stoken(TABLE, "table"),
			stoken(tBare, "foo"),
		},
	)

	testOK(
		"create table foo (col1, col2, col3)",
		[]token{
			stoken(CREATE, "create"),
			stoken(TABLE, "table"),
			stoken(tBare, "foo"),
			stoken('(', "("),
			stoken(tBare, "col1"),
			stoken(',', ","),
			stoken(tBare, "col2"),
			stoken(',', ","),
			stoken(tBare, "col3"),
			stoken(')', ")"),
		},
	)
	// *
	testOK(
		"select * from foo",
		[]token{
			stoken(SELECT, "select"),
			stoken(tOperator, "*"),
			stoken(FROM, "from"),
			stoken(tBare, "foo"),
		},
	)

	// fancy whitespace
	testOK(
		"  \tselect\n*\nfrom   foo ",
		[]token{
			stoken(SELECT, "select"),
			stoken(tOperator, "*"),
			stoken(FROM, "from"),
			stoken(tBare, "foo"),
		},
	)

	testOK(
		"from FROM 'from' ''",
		[]token{
			stoken(FROM, "from"),
			stoken(FROM, "FROM"),
			stoken(tLiteral, "from"),
			stoken(tLiteral, ""),
		},
	)

	testOK(
		"bare \"id 1\" [id 2] `id 3` 'lit 1'",
		[]token{
			stoken(tBare, "bare"),
			stoken(tIdentifier, "id 1"),
			stoken(tIdentifier, "id 2"),
			stoken(tIdentifier, "id 3"),
			stoken(tLiteral, "lit 1"),
		},
	)
	testOK(
		"|| * / % + - << >> & | < <= > >= = == != <> ~",
		[]token{
			stoken(tOperator, "||"),
			stoken(tOperator, "*"),
			stoken(tOperator, "/"),
			stoken(tOperator, "%"),
			stoken(tOperator, "+"),
			stoken(tOperator, "-"),
			stoken(tOperator, "<<"),
			stoken(tOperator, ">>"),
			stoken(tOperator, "&"),
			stoken(tOperator, "|"),
			stoken(tOperator, "<"),
			stoken(tOperator, "<="),
			stoken(tOperator, ">"),
			stoken(tOperator, ">="),
			stoken(tOperator, "="),
			stoken(tOperator, "=="),
			stoken(tOperator, "!="),
			stoken(tOperator, "<>"),
			stoken(tOperator, "~"),
		},
	)
	testOK(
		"IS NOT IN LIKE GLOB MATCH REGEXP AND OR",
		[]token{
			stoken(IS, "IS"),
			stoken(NOT, "NOT"),
			stoken(IN, "IN"),
			stoken(LIKE, "LIKE"),
			stoken(GLOB, "GLOB"),
			stoken(MATCH, "MATCH"),
			stoken(REGEXP, "REGEXP"),
			stoken(AND, "AND"),
			stoken(OR, "OR"),
		},
	)

	testError(
		"foo 'bar",
		errors.New("no terminating ' found"),
	)

}
