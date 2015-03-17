package gosession

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// session store provider list
var adapter = make(map[string]Provider)

// manger session store
// provider basic operation
//
const (
	SECURE           bool   = true
	COOKIENAME       string = "gosession"
	LIFETIME         int64  = 3600
	SESSIONIDLENGTH  int64  = 16
	COOKIEEXPIRES    int    = 3600
	COOKIENAMELENGHT int64  = 6
)

type SessionStore interface {
	Set(key, value interface{})
	Get(key interface{}) interface{}
	Delete(key interface{})
	SessionID() string
	Flush()
	All() map[interface{}]interface{}
}

// this interface for adapter
type Provider interface {
	InitConfig(gclifetime int64, config string) error
	CreateSession() (SessionStore, error)
	DestroySession(sid string) error
	GCSession()
}

type sessionConfig struct {
	CookieName        string //session name
	EnableSetCookie   bool
	Gctime            int64
	Expirestime       int64  //Expires times
	Secure            bool   //
	CookieExpirestime int    //
	Domain            string //host
	SessionIdLength   int64  //session id length
	ProviderConfig    string //provider config
}

// adapter session store style
type Adapter struct {
	store  Provider
	config *sessionConfig
}

// Registration session storage way, so that you can use has been registered and have implemented way of storage
func Register(provider string, store Provider) {
	if _, exist := adapter[provider]; !exist {
		adapter[provider] = store
	}
}

// Using the sha1 gave only calculate the string
func Secure(str string, length int64) string {
	str += COOKIENAME
	nstr := []byte(str)
	// sster := append(, []byte(SECURESTR))
	s := md5.New()
	newstr := s.Sum(nstr)
	return hex.EncodeToString(newstr[0:COOKIENAMELENGHT])
}

// auto collection session
// proivder must be have an SeesionGC function
func (adapter *Adapter) gc() {
	adapter.store.GCSession()
	// setting session garbage collection
	time.AfterFunc(time.Duration(adapter.config.Gctime)*time.Second, func() { adapter.gc() })
}

// Start seesion service
// The default is not started,Use session must be started manually
//
func (adapter *Adapter) StartSession(w http.ResponseWriter, r *http.Request) (store SessionStore, err error) {
	// Use the user-agent client value is encrypted to prevent session hijacked
	cookiename := Secure(r.Header.Get("User-Agent"), COOKIENAMELENGHT)
	cookie, errs := r.Cookie(cookiename)
	if errs != nil || cookie.Value == "" {
		store, err = adapter.store.CreateSession()
		sid := store.SessionID()
		if sid == "" {
			sid = Secure(sid, SESSIONIDLENGTH)
		}

		cookie = &http.Cookie{Name: cookiename,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   adapter.config.Secure,
			Domain:   r.Header.Get("Host")}
		if adapter.config.CookieExpirestime >= 0 {
			cookie.MaxAge = adapter.config.CookieExpirestime
		}
		if adapter.config.EnableSetCookie {
			http.SetCookie(w, cookie)
		}
		r.AddCookie(cookie)
	}
	return
}

// The session started at some point after must need to destroy
// Destroy session by its id in http request cookie.
func (adapter *Adapter) DestroySession(w http.ResponseWriter, r *http.Request) {
	cookiename := Secure(r.Header.Get("User-Agent")+adapter.config.CookieName, COOKIENAMELENGHT)
	cookie, err := r.Cookie(cookiename)
	if err != nil || cookie.Value == "" {
		return
	} else {
		adapter.store.DestroySession(cookie.Value)
		cookie := http.Cookie{Name: cookiename,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now(),
			MaxAge:   -1}
		http.SetCookie(w, &cookie)
	}
}

// The session is independent with the web framework, if used in other frameworks need through the bootstrap function configuration session
// example:
//      Sessions, _ = session.Bootstrap("file",`{"cookieName":"gosession","CookieExpirestime":3600,"ProviderConfig":"./tmp"}`)
// parmas: AdapterName, its adapter type name.
//         example : file ,it's using file store session data
// params: config, the config is the foundation of the session configuration and configuration of the adapters
//         It refer to the detailed configuration sessionConfig struct
func Bootstrap(adapterName, config string) (*Adapter, error) {
	provider, ok := adapter[adapterName]
	if !ok {
		return nil, fmt.Errorf("Not Found Session adapter %q (forgotten import?)", adapterName)
	}

	conf := new(sessionConfig)
	conf.EnableSetCookie = true
	err := json.Unmarshal([]byte(config), conf)
	if err != nil {
		return nil, err
	}
	if conf.CookieName == "" {
		conf.CookieName = COOKIENAME
	}
	if conf.Expirestime == 0 {
		conf.Expirestime = LIFETIME
	}
	if conf.Gctime == 0 {
		conf.Gctime = LIFETIME
	}
	if conf.Secure == false {
		conf.Secure = SECURE
	}
	if conf.CookieExpirestime == 0 {
		conf.CookieExpirestime = COOKIEEXPIRES
	}
	if conf.SessionIdLength == 0 {
		conf.SessionIdLength = SESSIONIDLENGTH
	}
	err = provider.InitConfig(conf.Expirestime, conf.ProviderConfig)
	if err != nil {
		return nil, err
	}
	adapter := &Adapter{provider, conf}
	go adapter.gc()
	return adapter, nil
}
