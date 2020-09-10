package middleware

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/gob"
	"encoding/json"
	gins "github.com/gin-contrib/sessions"
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

///////////////////////////////////////////////////////////////////////////////////////
// jsonSerializer 使用json对session数据进行序列化
type jsonSerializer struct{}

// Serialize json序列化
func (r jsonSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	return json.Marshal(s.Values)
}

// Deserialize json解序列化
func (r jsonSerializer) Deserialize(d []byte, s *sessions.Session) error {
	return json.Unmarshal(d, &s.Values)
}

///////////////////////////////////////////////////////////////////////////////////////
// gobSerializer 使用gob对session数据进行序列化
type gobSerializer struct{}

// Serialize gob序列化
func (r gobSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(s.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

// Deserialize gob解序列化
func (r gobSerializer) Deserialize(d []byte, s *sessions.Session) error {
	return gob.NewDecoder(bytes.NewBuffer(d)).Decode(&s.Values)
}

///////////////////////////////////////////////////////////////////////////////////////
// SSerializeTable session序列化的实现列表
var SSerializeTable = map[string]SessionSerializer{
	"json": jsonSerializer{},
	"gob":  gobSerializer{},
}

// Storage 用redis存储
type Storage struct {
	Redis      *redis.Client
	Codecs     []securecookie.Codec
	Opts       *sessions.Options
	Expire     int               // Session有效期
	keyPrefix  string            // Session名前缀
	serializer SessionSerializer // Session数据序列化
}

func NewStorageToRedis(cli *redis.Client, keyPairs ...[]byte) *Storage {
	return &Storage{
		Redis:  cli,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Opts: &sessions.Options{
			Path:   "/",
			MaxAge: 2592000,
		},
		Expire:     1440,
		keyPrefix:  "SESSION_",
		serializer: SSerializeTable["gob"],
	}
}

// SetSerializer 设置序列化方式
func (s *Storage) SetSerializer(sr SessionSerializer) {
	s.serializer = sr
}

// SetKeyPrefix 设置键前缀
func (s *Storage) SetKeyPrefix(prefix string) {
	if prefix == "" {
		s.keyPrefix = "SESSION_"
	} else {
		s.keyPrefix = prefix
	}
}

// SetSerializer 设置序列化方式
func (s *Storage) SetExpire(sr SessionSerializer) {
	s.serializer = sr
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
	options := *s.Opts
	ss.Options = &options
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
func (s *Storage) Save(_ *http.Request, w http.ResponseWriter, ss *sessions.Session) error {
	if ss.Options.MaxAge < 0 {
		if err := s.delete(ss); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(ss.Name(), "", ss.Options))
	} else {
		if ss.ID == "" {
			ss.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := s.save(ss); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(ss.Name(), ss.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(ss.Name(), encoded, ss.Options))
	}
	return nil
}

// MaxAge cookie的最长有效期
func (s *Storage) MaxAge(age int) {
	s.Opts.MaxAge = age
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// Options 选项
func (s *Storage) Options(opts gins.Options) {
	s.Opts = opts.ToGorillaOptions()
}

// save 存储数据
func (s *Storage) save(ss *sessions.Session) error {
	b, err := s.serializer.Serialize(ss)
	if err != nil {
		return err
	}
	return s.Redis.Do(context.Background(), "SETEX", s.keyPrefix+ss.ID, s.Expire, b).Err()
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
func (s *Storage) delete(ss *sessions.Session) error {
	return s.Redis.Del(context.Background(), s.keyPrefix+ss.ID).Err()
}
