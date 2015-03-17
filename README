# Go Session – Session in Golang


# Example:
Read the file：
    [example/test.go](example/test.go)

```go
    import (
        "github.com/wackonline/gosession"
    )
```

Then in you web app init the global session adapter,it like this

```go
    var provider *gosession.Adapter
    var session gosession.SessionStore
```

* Use **file** as provider, the last param is the path where you want file to be stored:

```go
    func init() {
        provider, _ = gosession.Bootstrap("file", `{"cookieName":"gosessionid","Gctime":3600,"ProviderConfig":"./tmp"}`)
        session, _ = provider.StartSession(w, r)
    }
```

Finally in the handlerfunc you can use it like this:

```go
    func say(w http.ResponseWriter, r *http.Request) {

        session.Set("joe", "hello world!...")
        joe := session.Get("joe")
        fmt.Fprintln(w, joe)
        fmt.Fprintln(w, "================")
    }

```

## How to write own provider?
```go
    type SessionStore interface {
        Set(key, value interface{}) // like as session.Set(Key,Value)
        Get(key interface{}) interface{} // like as session.Get(Key) ==> value 
        Delete(key interface{}) // like as session.Delete(Key) ===> remove map[interface{}]interface{} index for Key's data
        SessionID() string // return SessionId,this id for start session created
        Flush() // delete all session data
        All() map[interface{}]interface{} // get all session data
    }

    // this interface for adapter
    type Provider interface {
        InitConfig(gclifetime int64, config string) error // init provider config
        CreateSession(sid string) (SessionStore, error) // Create session by provider,such as **file**,it's mkdir path.
        DestroySession(sid string) error // destroy session,such as user logout,destroy user session
        GCSession() // Automatic collection treatment session date
    }
```

# Refer
    [golang manual][http://golang.org]
    [beego framework](http://beego.me)-[github](https://github.com/astaxie/beego)
    [build web application with golang](https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/06.1.md)
    [martini-contrib/sessions](https://github.com/martini-contrib/sessions)

# License

Structre Record is released under the GPLV3 license:
    [License](https://github.com/wackonline/structrecord/blob/master/LICENSE)