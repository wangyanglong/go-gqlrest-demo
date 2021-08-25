package graph

import (
	"fmt"
	"testing"
	"time"

	gqlcli "github.com/99designs/gqlgen/client"
	restcli "github.com/speedoops/gql2rest/client"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlgen-todos/graph/utils/mock"
)

const T9527 = "T9527" // 这是一个特定的值，第一个创建的 Todo 的 ID

func getTestString() string {
	return fmt.Sprintf("text_%s", time.Now().Format("2006-01-02 15:04:05"))
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	User User   `json:"user"`
}

func TestTodo(t *testing.T) {
	srv := mock.NewGraphQLServer(&Resolver{})
	c := gqlcli.New(srv, gqlcli.Path("/query"))

	t.Run("mutation.createTodo", func(t *testing.T) {
		var resp struct {
			CreateTodo Todo `json:"createTodo"`
		}

		mutation := `
		mutation createTodo($text:String!) {
			createTodo(input: {userID:"uid", text:$text}){
			  id,text,done
			}
		  }
		`
		c.MustPost(mutation, &resp, gqlcli.Var("text", getTestString()))
		c.MustPost(mutation, &resp, gqlcli.Var("text", getTestString()))
		c.MustPost(mutation, &resp, gqlcli.Var("text", getTestString()))

		require.NotEmpty(t, resp.CreateTodo)
		t.Logf("%+v", resp.CreateTodo)
	})

	t.Run("mutation.updateTodo", func(t *testing.T) {
		var resp struct {
			UpdateTodo Todo `json:"updateTodo"`
		}

		mutation := `
		mutation updateTodo($id: ID!) {
			updateTodo(id: $id, input: {userID:"uid", text:"9527.Updated"}){
			  id,text,done
			}
		  }
		`
		c.MustPost(mutation, &resp, gqlcli.Var("id", T9527))

		require.NotEmpty(t, resp.UpdateTodo)
		t.Logf("%+v", resp.UpdateTodo)
	})

	t.Run("query.todos should contain T9527", func(t *testing.T) {
		var resp struct {
			Todos []Todo `json:"todos"`
		}
		query := `
		query todos {
			todos(ids:["T9527"],userId2:"userID2",text2:"text2",done2:true) {
				id,text,done
			}
		}
		`
		c.MustPost(query, &resp)

		require.NotEmpty(t, resp.Todos)
		for _, v := range resp.Todos {
			if v.ID == T9527 {
				return
			}
		}
		require.Fail(t, "T9527 not found")
		t.Logf("%+v", resp.Todos)
	})

	t.Run("mutation.deleteTodo", func(t *testing.T) {
		var resp struct {
			DeleteTodo bool `json:"deleteTodo"`
		}

		mutation := `
		mutation deleteTodo($id: ID!) {
			deleteTodo(id: $id)
		  }
		`
		c.MustPost(mutation, &resp, gqlcli.Var("id", T9527))

		require.NotEmpty(t, resp.DeleteTodo)
		t.Logf("%+v", resp.DeleteTodo)
	})

	t.Run("query.todos should not contain T9527", func(t *testing.T) {
		var resp struct {
			Todos []Todo `json:"todos"`
		}
		query := `
		query todos {
			todos(ids:["T9527"],userId2:"userID2",text2:"text2",done2:true) {
				id,text,done
			}
		}
		`
		c.MustPost(query, &resp)

		require.NotEmpty(t, resp.Todos)
		for _, v := range resp.Todos {
			require.NotEqual(t, v.ID, T9527)
		}
		t.Logf("%+v", resp.Todos)
	})
}

// https://www.ontestautomation.com/an-introduction-to-rest-api-testing-in-go-with-resty/
func TestTodos_REST(t *testing.T) {

	s := mock.NewGraphQLServer(&Resolver{})
	c := restcli.New(s, restcli.Prefix("/api/v1"))

	t.Run("rest.createTodo", func(t *testing.T) {
		var resp struct {
			Todo Todo `json:"createTodo"`
		}

		payload := `
		{"input": {"userID":"uid", "text":"$text"}}
		`
		err := c.Post("/todo", &resp, restcli.Body(payload))
		require.Nil(t, err)
		require.NotEmpty(t, resp.Todo)

		t.Logf("%+v", resp.Todo)
	})

	t.Run("rest.updateTodo", func(t *testing.T) {
		var resp struct {
			Todo Todo `json:"updateTodo"`
		}

		payload := `
		{"input": {"userID":"uid", "text":"$text.Updated"}}
		`
		err := c.Put("/todo/T9527", &resp, restcli.Body(payload))
		require.Nil(t, err)
		require.NotEmpty(t, resp.Todo)

		t.Logf("%+v", resp.Todo)
	})

	t.Run("rest.deleteTodo", func(t *testing.T) {
		var resp struct {
			Todo bool `json:"deleteTodo"`
		}

		err := c.Delete("/todo/T9527", &resp)
		require.Nil(t, err)
		require.NotEmpty(t, resp.Todo)

		t.Logf("%+v", resp.Todo)
	})

	t.Run("rest.todos", func(t *testing.T) {
		var resp struct {
			Todos []Todo `json:"todos"`
		}

		err := c.Get("/todos?ids=T9527&userId2=userId2&text2=text2&done2=true", &resp)
		require.Nil(t, err)

		for _, v := range resp.Todos {
			if v.ID == T9527 {
				t.Fail()
			}
		}
		t.Logf("%+v", resp.Todos)
	})
}
