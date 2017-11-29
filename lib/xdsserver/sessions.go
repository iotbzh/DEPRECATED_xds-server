package xdsserver

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	uuid "github.com/satori/go.uuid"
	"github.com/syncthing/syncthing/lib/sync"
)

const sessionCookieName = "xds-sid"
const sessionHeaderName = "XDS-SID"

const sessionMonitorTime = 10 // Time (in seconds) to schedule monitoring session tasks

const initSessionMaxAge = 10 // Initial session max age in seconds
const maxSessions = 100000   // Maximum number of sessions in sessMap map

const secureCookie = false // TODO: see https://github.com/astaxie/beego/blob/master/session/session.go#L218

// ClientSession contains the info of a user/client session
type ClientSession struct {
	ID       string
	WSID     string // only one WebSocket per client/session
	MaxAge   int64
	IOSocket *socketio.Socket

	// private
	expireAt time.Time
	useCount int64
}

// Sessions holds client sessions
type Sessions struct {
	*Context
	cookieMaxAge int64
	sessMap      map[string]ClientSession
	mutex        sync.Mutex
	stop         chan struct{} // signals intentional stop
}

// NewClientSessions .
func NewClientSessions(ctx *Context, cookieMaxAge string) *Sessions {
	ckMaxAge, err := strconv.ParseInt(cookieMaxAge, 10, 0)
	if err != nil {
		ckMaxAge = 0
	}
	s := Sessions{
		Context:      ctx,
		cookieMaxAge: ckMaxAge,
		sessMap:      make(map[string]ClientSession),
		mutex:        sync.NewMutex(),
		stop:         make(chan struct{}),
	}
	s.WWWServer.router.Use(s.Middleware())

	// Start monitoring of sessions Map (use to manage expiration and cleanup)
	go s.monitorSessMap()

	return &s
}

// Stop sessions management
func (s *Sessions) Stop() {
	close(s.stop)
}

// Middleware is used to managed session
func (s *Sessions) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// FIXME Add CSRF management

		// Get session
		sess := s.Get(c)
		if sess == nil {
			// Allocate a new session key and put in cookie
			sess = s.newSession("")
		} else {
			s.refresh(sess.ID)
		}

		// Set session in cookie and in header
		// Do not set Domain to localhost (http://stackoverflow.com/questions/1134290/cookies-on-localhost-with-explicit-domain)
		c.SetCookie(sessionCookieName, sess.ID, int(sess.MaxAge), "/", "",
			secureCookie, false)
		c.Header(sessionHeaderName, sess.ID)

		// Save session id in gin metadata
		c.Set(sessionCookieName, sess.ID)

		c.Next()
	}
}

// Get returns the client session for a specific ID
func (s *Sessions) Get(c *gin.Context) *ClientSession {
	var sid string

	// First get from gin metadata
	v, exist := c.Get(sessionCookieName)
	if v != nil {
		sid = v.(string)
	}

	// Then look in cookie
	if !exist || sid == "" {
		sid, _ = c.Cookie(sessionCookieName)
	}

	// Then look in Header
	if sid == "" {
		sid = c.Request.Header.Get(sessionCookieName)
	}
	if sid != "" {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		if key, ok := s.sessMap[sid]; ok {
			// TODO: return a copy ???
			return &key
		}
	}
	return nil
}

// IOSocketGet Get socketio definition from sid
func (s *Sessions) IOSocketGet(sid string) *socketio.Socket {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess, ok := s.sessMap[sid]
	if ok {
		return sess.IOSocket
	}
	return nil
}

// UpdateIOSocket updates the IO Socket definition for of a session
func (s *Sessions) UpdateIOSocket(sid string, so *socketio.Socket) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.sessMap[sid]; ok {
		sess := s.sessMap[sid]
		if so == nil {
			// Could be the case when socketio is closed/disconnected
			sess.WSID = ""
		} else {
			sess.WSID = (*so).Id()
		}
		sess.IOSocket = so
		s.sessMap[sid] = sess
	}
	return nil
}

// nesSession Allocate a new client session
func (s *Sessions) newSession(prefix string) *ClientSession {
	uuid := prefix + uuid.NewV4().String()
	id := base64.URLEncoding.EncodeToString([]byte(uuid))
	se := ClientSession{
		ID:       id,
		WSID:     "",
		MaxAge:   initSessionMaxAge,
		IOSocket: nil,
		expireAt: time.Now().Add(time.Duration(initSessionMaxAge) * time.Second),
		useCount: 0,
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessMap[se.ID] = se

	s.Log.Debugf("NEW session (%d): %s", len(s.sessMap), id)
	return &se
}

// refresh Move this session ID to the head of the list
func (s *Sessions) refresh(sid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sess := s.sessMap[sid]
	sess.useCount++
	if sess.MaxAge < s.cookieMaxAge && sess.useCount > 1 {
		sess.MaxAge = s.cookieMaxAge
		sess.expireAt = time.Now().Add(time.Duration(sess.MaxAge) * time.Second)
	}

	// TODO - Add flood detection (like limit_req of nginx)
	// (delayed request when to much requests in a short period of time)

	s.sessMap[sid] = sess
}

func (s *Sessions) monitorSessMap() {
	for {
		select {
		case <-s.stop:
			s.Log.Debugln("Stop monitorSessMap")
			return
		case <-time.After(sessionMonitorTime * time.Second):
			if s.LogLevelSilly {
				s.Log.Debugf("Sessions Map size: %d", len(s.sessMap))
				s.Log.Debugf("Sessions Map : %v", s.sessMap)
			}

			if len(s.sessMap) > maxSessions {
				s.Log.Errorln("TOO MUCH sessions, cleanup old ones !")
			}

			s.mutex.Lock()
			for _, ss := range s.sessMap {
				if ss.expireAt.Sub(time.Now()) < 0 {
					s.Log.Debugf("Delete expired session id: %s", ss.ID)
					delete(s.sessMap, ss.ID)
				}
			}
			s.mutex.Unlock()
		}
	}
}
