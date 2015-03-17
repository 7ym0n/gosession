package gosession

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	// "fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

var (
	fileProvder   = &FileProvider{}
	gcmaxlifetime int64
	md5values     string
)

// File session store
type FileSessionStore struct {
	sid    string //session id
	lock   sync.RWMutex
	values map[interface{}]interface{} // session key/value
}

// File session provider
type FileProvider struct {
	lock        sync.RWMutex
	expirestime int64
	savePath    string // session saved path
}

// Set value to file session
func (fs *FileSessionStore) Set(key, value interface{}) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	var tmp map[interface{}]interface{}
	tmp = fs.values
	tmp[key] = value
	if b := fileProvder.isWriteFile(tmp); b {
		fs.values[key] = value
		fs.sid = fileProvder.sessionFileName(fileProvder.encode(fs.values))
		fileProvder.writeFile(path.Join(fileProvder.savePath, fs.sid), fs.values)

	}

}

// Get value from file session
func (fs *FileSessionStore) Get(key interface{}) interface{} {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	if v, ok := fs.values[key]; ok {
		return v
	} else {
		p := path.Join(fileProvder.savePath, fs.sid)
		if b := fileProvder.isFile(p); b {
			result, err := fileProvder.readFile(p)
			if err == nil {
				if v, ok := result[key]; ok {
					return v
				}
			}
		}
		return nil
	}
}

// Delete value in file session by given key
func (fs *FileSessionStore) Delete(key interface{}) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	var tmp map[interface{}]interface{}
	tmp = fs.values
	delete(tmp, key)
	if b := fileProvder.isWriteFile(tmp); b {
		delete(fs.values, key)
		fs.sid = fileProvder.sessionFileName(fileProvder.encode(fs.values))
		fileProvder.writeFile(path.Join(fileProvder.savePath, fs.sid), fs.values)
	}

}

// Clean all values in file session
func (fs *FileSessionStore) Flush() {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values = make(map[interface{}]interface{})
	fs.sid = fileProvder.sessionFileName(fileProvder.encode(fs.values))
}

// return all in file session
func (fs *FileSessionStore) All() map[interface{}]interface{} {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	return fs.values
}

// Get file session store id
func (fs *FileSessionStore) SessionID() string {
	return fs.sid
}

// Init file session provider.
// savePath sets the session files path.
func (fp *FileProvider) InitConfig(expirestime int64, savePath string) error {
	fp.expirestime = expirestime
	fp.savePath = savePath
	return nil
}

// Read file session by sid.
// if file is not exist, create it.
// the file path is generated from sid string.
// func (fp *FileProvider) ReadSession(sid string) (SessionStore, error) {

// }

// create an session
func (fp *FileProvider) CreateSession() (SessionStore, error) {

	fileProvder.lock.Lock()
	defer fileProvder.lock.Unlock()
	//var kv map[interface{}]interface{}
	kv := make(map[interface{}]interface{})
	session := &FileSessionStore{sid: "", values: kv}
	return session, os.MkdirAll(fp.savePath, 0655)
}

// Remove all files in this save path
func (fp *FileProvider) DestroySession(sid string) error {
	fileProvder.lock.Lock()
	defer fileProvder.lock.Unlock()
	os.Remove(path.Join(fp.savePath, sid))
	return nil
}

func (fp *FileProvider) GCSession() {
	fileProvder.lock.Lock()
	defer fileProvder.lock.Unlock()
	filepath.Walk(fp.savePath, fp.clearFile)
}

// remove file in save path if expired
func (fp *FileProvider) clearFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if (info.ModTime().Unix() + fp.expirestime) < time.Now().Unix() {
		os.Remove(path)
	}
	return nil
}

func (fp *FileProvider) isFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// read session file
func (fp *FileProvider) readFile(path string) (map[interface{}]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err == nil {
		result, errs := fp.decode(data)
		if errs == nil {
			return result, nil
		}

	}
	result := make(map[interface{}]interface{})
	return result, err
}

// write session data to file
func (fp *FileProvider) writeFile(path string, data map[interface{}]interface{}) {
	ioutil.WriteFile(path, fp.encode(data), 0655)
}

// Assert whether to write the data
func (fp *FileProvider) isWriteFile(kv map[interface{}]interface{}) bool {
	// s := md5.New()
	// s.Write([]byte(fp.encode(kv)))
	// newstr := hex.EncodeToString(s.Sum(nil))
	err := fp.isFile(path.Join(fileProvder.savePath, fp.sessionFileName(fileProvder.encode(kv))))
	//Compare the new value and old values are equal
	if !err {
		// md5values = newstr
		// If err equals nil , MD5 value is different

		return true
	}
	return false

}

//Use md5 generate session file name
func (fp *FileProvider) sessionFileName(str []byte) string {
	s := md5.New()
	return hex.EncodeToString(s.Sum(str[15:20]))
}

//The structure of map encryption
func (fp *FileProvider) encode(maps map[interface{}]interface{}) []byte {
	for _, v := range maps {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(maps)
	if err != nil {
		return []byte("")
	}
	return buf.Bytes()
}

//The structure of map decryption
func (fp *FileProvider) decode(encoded []byte) (map[interface{}]interface{}, error) {
	buffer := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buffer)
	var out map[interface{}]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func init() {
	Register("file", fileProvder)
}
