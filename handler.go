package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")

	callID := r.FormValue("CallSid")
	retryKey, errorKey := callID+"-retry", callID+"-error"

	exitData := struct {
		Name  string
		Phone string
	}{
		s.config.Name,
		s.config.Phone,
	}

	retryCount, err := s.getValue(retryKey)
	if err != nil {
		s.log.Infof("%v: failed to get redis key: %v", callID, err)
		s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
		return
	}

	errorCount, err := s.getValue(errorKey)
	if err != nil {
		s.log.Infof("%v: failed to get redis key: %v", callID, err)
		s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
		return
	}

	recordingURL := r.FormValue("RecordingUrl")
	if recordingURL == "" {
		if err := s.redis.Incr(retryKey).Err(); err != nil {
			s.log.Infof("%v: failed to increment redis key: %v", callID, err)
			s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
			return
		}
		s.log.Infof("%v: playing welcome message", callID)
		s.template.ExecuteTemplate(w, "welcome.xml", s.config.Timeout)
		return
	}

	b, err := fetchAudio(recordingURL)
	if err != nil {
		if errorCount < s.config.MaxErrors-1 {
			if err := s.redis.Incr(errorKey).Err(); err != nil {
				s.log.Infof("%v: failed to increment redis key: %v", callID, err)
				s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
				return
			}
			s.log.Infof("%v: failed to fetch audio: %v", callID, err)
			s.template.ExecuteTemplate(w, "error.xml", s.config.Timeout)
			return
		}
		s.log.Infof("%v: maximum errors reached", callID)
		s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
		return
	}

	text, err := fetchTranscript(b, s.config.APIKeyGoogle)
	if err != nil {
		if errorCount < s.config.MaxErrors-1 {
			if err := s.redis.Incr(errorKey).Err(); err != nil {
				s.log.Infof("%v: failed to increment redis key: %v", callID, err)
				s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
				return
			}
			s.log.Infof("%v: failed to fetch transcript: %v", callID, err)
			s.template.ExecuteTemplate(w, "error.xml", s.config.Timeout)
			return
		}
		s.log.Infof("%v: maximum errors reached", callID)
		s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
		return
	}

	if strings.Contains(text, s.config.Password) {
		s.log.Infof("%v: allowed into apartment", callID)
		s.template.ExecuteTemplate(w, "success.xml", s.config.EntryDigit)
		return
	}

	if retryCount <= s.config.MaxRetries {
		if err := s.redis.Incr(retryKey).Err(); err != nil {
			s.log.Infof("%v: failed to increment redis key: %v", callID, err)
			s.template.ExecuteTemplate(w, "fatalerror.xml", exitData)
			return
		}
		s.log.Infof("%v: retrying; retry %v of %v", callID, retryCount, s.config.MaxRetries)
		data := struct {
			AttemptsRemaining int
			Timeout           int
		}{
			s.config.MaxRetries - retryCount + 1,
			s.config.Timeout,
		}
		s.template.ExecuteTemplate(w, "retry.xml", data)
		return
	}
	s.log.Infof("%v: failed to provide magic word; forwarding to phone", callID)
	s.template.ExecuteTemplate(w, "failure.xml", exitData)
	return
}

func (s *server) getValue(key string) (int, error) {
	r, err := s.redis.Get(key).Result()
	if err != nil {
		if err = s.redis.Set(key, 0, time.Second*60*60).Err(); err != nil {
			return 0, err
		}
		r = "0"
	}
	v, err := strconv.Atoi(r)
	if err != nil {
		return 0, err
	}
	return v, nil
}
