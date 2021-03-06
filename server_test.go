package wine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gopub/wine"
)

var server *wine.Server

type testJSONObj struct {
	Name string
	Age  int
}

func TestMain(m *testing.M) {
	server = wine.NewServer(wine.DefaultConfig())
	go func() {
		server.Run(":8000")
	}()

	time.Sleep(time.Second)
	result := m.Run()
	os.Exit(result)
}

func TestJSON(t *testing.T) {
	obj := &testJSONObj{
		Name: "tom",
		Age:  19,
	}
	server.Get("/json", func(ctx context.Context, _ *wine.Request, _ wine.Invoker) wine.Responsible {
		return wine.JSON(http.StatusOK, obj)
	})

	resp, err := http.DefaultClient.Get("http://localhost:8000/json")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var result testJSONObj
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Log(string(data))
		t.Fatal(err)
	}

	if result != *obj {
		t.Fatal(result, *obj)
	}

	if resp.Header[wine.ContentType][0] != wine.MIMEJSON {
		t.Fatal(resp.Header[wine.ContentType])
	}
}

func TestHTML(t *testing.T) {
	var htmlText = `
	<html>
		<header>
		</header>
		<body>
			Hello, world!
		</body>
	</html>
	`
	server.Get("/html/hello.html", func(ctx context.Context, _ *wine.Request, _ wine.Invoker) wine.Responsible {
		return wine.HTML(http.StatusOK, htmlText)
	})

	resp, err := http.DefaultClient.Get("http://localhost:8000/html/hello.html")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if string(data) != htmlText {
		t.Fatal(string(data))
	}

	if resp.Header[wine.ContentType][0] != wine.MIMEHTML {
		t.Fatal(resp.Header[wine.ContentType])
	}
}

func TestPathParams(t *testing.T) {
	server.Get("/sum/{a}/{b}", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		a := req.Parameters.Int("a")
		b := req.Parameters.Int("b")
		return wine.Text(http.StatusOK, fmt.Sprint(a+b))
	})

	{
		resp, err := http.DefaultClient.Get("http://localhost:8000/sum/1/2")
		if err != nil {
			t.Fatal(err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if string(data) != "3" {
			t.Fatal(string(data))
		}
	}

	{
		resp, err := http.DefaultClient.Get("http://localhost:8000/sum/1")
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(resp.Status)
		}
	}
}
