package middleware

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/gob"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

// SessionSerializer provides an interface hook for alternative serializers
type SessionSerializer interface {
	Deserialize(d []byte, s *sessions.Session) error
	Serialize(s *sessions.Session) ([]byte, error)
}

// JSONSerializer 使用json对session数据进行序列化
type JSONSerializer struct{}

// Serialize json序列化
func (r JSONSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	return json.Marshal(s.Values)
}

// Deserialize json解序列化
func (r JSONSerializer) Deserialize(d []byte, s *sessions.Session) error {
	return json.Unmarshal(d, &s.Values)
}

// GobSerializer 使用gob对session数据进行序列化
type GobSerializer struct{}

// Serialize gob序列化
func (r GobSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(s.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

// Deserialize gob解序列化
func (r GobSerializer) Deserialize(d []byte, s *sessions.Session) error {
	return gob.NewDecoder(bytes.NewBuffer(d)).Decode(&s.Values)
}

// Storage 用redis存储
type Storage struct {
	Redis      *redis.Client
	Codecs     []securecookie.Codec
	Options    *sessions.Options
	Expire     int               // Session有效期
	keyPrefix  string            // Session名前缀
	serializer SessionSerializer // Session数据序列化
}

func NewStorageToRedis(cli *redis.Client, keyPairs ...[]byte) *Storage {
	return &Storage{
		Redis:  cli,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 2592000,
		},
		Expire:     1800,
		keyPrefix:  "session_",
		serializer: GobSerializer{},
	}
}

// Close 关闭Redis连接
func (s *Storage) Close() error {
	return s.Redis.Close()
}

// Get 注册会话
func (s *Storage) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New
func (s *Storage) New(r *http.Request, name string) (*sessions.Session, error) {
	var (
		err error
		ok  bool
	)
	ss := sessions.NewSession(s, name)
	//// make a copy
	//options := *s.Options
	//session.Options = &options
	ss.IsNew = true
	if c, err := r.Cookie(name); err == nil {
		err = securecookie.DecodeMulti(name, c.Value, &ss.ID, s.Codecs...)
		if err == nil {
			ok, err = s.load(ss)
			ss.IsNew = !(err == nil && ok) // not new if no error and data available
		}
	}
	return ss, err
}

// Save adds a single session to the response.
func (s *Storage) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Marked for deletion.
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	} else {
		// Build an alphanumeric key for the redis store.
		if session.ID == "" {
			session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := s.save(session); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	}
	return nil
}

// save stores the session in redis.
func (s *Storage) save(ss *sessions.Session) error {
	b, err := s.serializer.Serialize(ss)
	if err != nil {
		return err
	}
	age := ss.Options.MaxAge
	if age == 0 {
		age = s.Expire
	}
	return s.Redis.Do(context.Background(), "SETEX", s.keyPrefix+ss.ID, age, b).Err()
}

// load 加载数据
// 从redis中加载session数据，如果成功返回 true
func (s *Storage) load(ss *sessions.Session) (bool, error) {
	data, err := s.Redis.Get(context.Background(), s.keyPrefix+ss.ID).Bytes()
	if err != nil {
		return false, err
	}
	return true, s.serializer.Deserialize(data, ss)
}

// delete 删除数据
func (s *Storage) delete(session *sessions.Session) error {
	return s.Redis.Del(context.Background(), s.keyPrefix+session.ID).Err()
}
