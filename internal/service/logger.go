package service

import (
	"sync"
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

type RequestLogger struct {
	db        *gorm.DB
	ch        chan *model.RequestLog
	flushSize int
	stop      chan struct{}
	once      sync.Once
}

func NewRequestLogger(db *gorm.DB, bufferSize, flushSize int) *RequestLogger {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	if flushSize <= 0 {
		flushSize = 100
	}
	l := &RequestLogger{
		db:        db,
		ch:        make(chan *model.RequestLog, bufferSize),
		flushSize: flushSize,
		stop:      make(chan struct{}),
	}
	go l.run()
	return l
}

func (l *RequestLogger) Log(entry *model.RequestLog) {
	if entry == nil {
		return
	}
	select {
	case l.ch <- entry:
	default:
	}
}

func (l *RequestLogger) Close() {
	l.once.Do(func() {
		close(l.stop)
	})
}

func (l *RequestLogger) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	batch := make([]*model.RequestLog, 0, l.flushSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = l.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}
			return nil
		})
		batch = batch[:0]
	}

	for {
		select {
		case <-l.stop:
			flush()
			return
		case e := <-l.ch:
			batch = append(batch, e)
			if len(batch) >= l.flushSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
