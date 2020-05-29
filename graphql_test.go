package graphql_test

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardanlabs/graphql"
	"github.com/google/go-cmp/cmp"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestGraphQL validates all the client support.
func TestGraphQL(t *testing.T) {
	t.Run("query", query)
	t.Run("error", errors)
}

func query(t *testing.T) {
	type document struct {
		Field1 string  `json:"field1"`
		Field2 int     `json:"field2"`
		Field3 float64 `json:"field3"`
		Field4 bool    `json:"field4"`
	}

	type response struct {
		Documents []document `json:"documents"`
	}

	var queryString = `query { getCity(id: "0x01") { id name lat lng } }`
	var clientString = `{"query":"query { getCity(id: \"0x01\") { id name lat lng } }","variables":{"key1":10,"key2":"hello","key3":28.45}}` + "\n"

	t.Log("Given the need to be able to validate processing a query.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a basic query: %s", testID, queryString)
		{
			f := func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(r.Method, http.MethodPost); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould see this is a POST call. Diff:\n%s", failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould see this is a POST call.", success, testID)

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to read the body: %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to read the body.", success, testID)

				if diff := cmp.Diff(string(b), clientString); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get the expected query. Diff:\n%s", failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get the expected query.", success, testID)

				io.WriteString(w, `{
					"data": {
						"documents": [
							{
								"field1": "a",
								"field2": 1,
								"field3": 3.14,
								"field4": true
							}
						]
					}
				}`)
			}

			server := httptest.NewServer(http.HandlerFunc(f))
			defer server.Close()

			gql := graphql.New(graphql.HTTP, server.URL[7:], http.DefaultClient)

			queryVars := map[string]interface{}{"key1": 10, "key2": "hello", "key3": 28.45}
			var got response
			if err := gql.QueryWithVars(context.Background(), graphql.CmdQuery, queryString, queryVars, &got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to execute the query: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to execute the query.", success, testID)

			exp := response{
				Documents: []document{
					{Field1: "a", Field2: 1, Field3: 3.14, Field4: true},
				},
			}

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", success, testID)
		}
	}
}

func errors(t *testing.T) {
	type document struct {
		Field1 string  `json:"field1"`
		Field2 int     `json:"field2"`
		Field3 float64 `json:"field3"`
		Field4 bool    `json:"field4"`
	}

	type response struct {
		Documents []document `json:"documents"`
	}

	var queryString = `query { getCity(id: "0x01") { id name lat lng } }`
	var clientString = `{"query":"query { getCity(id: \"0x01\") { id name lat lng } }","variables":null}` + "\n"

	t.Log("Given the need to be able to validate process a query with error.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a basic query: %s", testID, queryString)
		{
			f := func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(r.Method, http.MethodPost); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould see this is a POST call. Diff:\n%s", failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould see this is a POST call.", success, testID)

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to read the body: %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to read the body.", success, testID)

				if diff := cmp.Diff(string(b), clientString); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get the expected query. Diff:\n%s", failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get the expected query.", success, testID)

				io.WriteString(w, `{
					"errors": [
						{
							"message": "error forced by test"
						}
					]
				}`)
			}

			server := httptest.NewServer(http.HandlerFunc(f))
			defer server.Close()

			gql := graphql.New(graphql.HTTP, server.URL[7:], http.DefaultClient)

			var got response
			err := gql.Query(context.Background(), queryString, &got)
			if err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to execute the query with error.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to execute the query with error.", success, testID)
		}
	}
}
