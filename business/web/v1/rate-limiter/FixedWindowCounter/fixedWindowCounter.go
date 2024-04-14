package fixedwindowcounter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Zanda256/rate-limiter-go/business/data/cache"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

const RoundedToSeconds = "2006-01-02 15:04:05"

type WindowController struct {
	Log        *logger.Logger
	Store      *cache.RedisCache
	WindowSize int64
	MaxTokens  int
}

type WindowControllerConfig struct {
	Log        *logger.Logger
	Store      *cache.RedisCache
	WindowSize int64
	MaxTokens  int
}

func NewWindowController(cfg WindowControllerConfig) *WindowController {
	return &WindowController{
		Log:        cfg.Log,
		Store:      cfg.Store,
		WindowSize: cfg.WindowSize,
		MaxTokens:  cfg.MaxTokens,
	}
}

type Window struct {
	UserID      string `json:"userId"`
	CreatedAt   int64  `json:"createdAt"`
	MaxRequests int    `json:"maxRequests"`
	Requests    int    `json:"requests"`
}

func (w Window) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(w)
	return
}

func UnmarshalBinarytoWindow(data []byte, t *Window) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("unmarshal to TokenBucket{} failed %+v", err.Error()))
	}
	return nil
}

type WindowConfig struct {
	UserID     string
	WindowSize int64
	MaxTokens  int
}

func (wc *WindowController) NewWindow(cfg WindowConfig) Window {
	nowUnix := time.Now().Unix()
	wID := nowUnix / wc.WindowSize

	return Window{
		UserID:      cfg.UserID,
		CreatedAt:   wID,
		MaxRequests: wc.MaxTokens,
		Requests:    0,
	}
}

func (wc *WindowController) Accept(userID string) bool {
	v, err := wc.getWindow(userID)
	if err != nil {
		// If the key isn't present, create a new one
		if errors.Is(err, cache.ErrKeyNotFound) { // move to string constatnt later
			newWnd := wc.NewWindow(WindowConfig{
				UserID:     userID,
				WindowSize: wc.WindowSize,
				MaxTokens:  wc.MaxTokens,
			})
			if _, err = wc.updateRequests(newWnd); err != nil {
				wc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
				return false
			}
			wc.Log.Info(context.Background(), "successfully stored new value", newWnd)
			//If no error accept the request
			return true
		}
		// Other error type we log it and return false
		wc.Log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
		return false
	}
	theWindow := Window{}
	var (
		windowInBytes string
		ok            bool
	)

	if windowInBytes, ok = v.(string); !ok {
		wc.Log.Info(context.Background(), "cannot marshal retrieved value into string")
		return false
	}

	if err = UnmarshalBinarytoWindow([]byte(windowInBytes), &theWindow); err != nil {
		wc.Log.Info(context.Background(), "cannot marshal retrieved value into Window")
		return false
	}

	nowUnix := time.Now().Unix()
	currentWindow := nowUnix / wc.WindowSize

	if currentWindow == theWindow.CreatedAt {
		// still in current time window, check availability of requests
		if theWindow.Requests >= theWindow.MaxRequests {
			return false
		}
		// valid request, update window object
		if _, err = wc.updateRequests(theWindow); err != nil {
			wc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
			return false
		}
		return true
	}
	// create new window object
	newWnd := wc.NewWindow(WindowConfig{
		UserID:     userID,
		WindowSize: wc.WindowSize,
		MaxTokens:  wc.MaxTokens,
	})
	if _, err = wc.updateRequests(newWnd); err != nil {
		wc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
		return false
	}
	return true
}

func (bc *WindowController) getWindow(UserID string) (any, error) {
	v, err := bc.Store.RetrieveValue(context.Background(), UserID)
	if err != nil {
		bc.Log.Error(context.Background(), "checkbucket: %s", err.Error())
		return nil, err
	}
	if v == nil {
		bc.Log.Warn(context.Background(), "key not found")
		return nil, cache.ErrKeyNotFound
	}
	bc.Log.Info(context.Background(), "bucket in check bucket: %+v", v)
	return v, nil
}

func (bc *WindowController) updateRequests(w Window) (Window, error) {
	// increment requests by 1 and persist the result. If successful,accept and process the request.
	w.Requests += 1
	res, err := bc.Store.StoreValue(context.Background(), w.UserID, w, 30)
	if err != nil {
		bc.Log.Error(context.Background(), fmt.Sprintf("Store window value failed: %s", err.Error()))
		return Window{}, err
	}
	w = res.(Window)
	return w, nil
}
