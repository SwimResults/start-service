package service

// Live state per meeting (collection "live"): kept in memory, loaded lazily from Mongo, and written back.
// Timing data arrives every few seconds, so last_data_at is persisted at most every livePersistInterval;
// message changes are persisted at once. start-service runs as a single replica, so memory is the truth.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const livePersistInterval = 10 * time.Second

var liveCollection *mongo.Collection

func liveService(database *mongo.Database) {
	liveCollection = database.Collection("live")
}

type liveEntry struct {
	live        model.Live
	persistedAt time.Time // last write of last_data_at
}

var (
	liveMu      sync.Mutex
	liveEntries = map[string]*liveEntry{}
)

// liveEntryLocked returns the entry of a meeting, loading it from Mongo on first use. Callers hold liveMu.
func liveEntryLocked(meeting string) *liveEntry {
	if e, ok := liveEntries[meeting]; ok {
		return e
	}
	e := &liveEntry{live: model.Live{Meeting: meeting}}
	if liveCollection != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var stored model.Live
		err := liveCollection.FindOne(ctx, bson.M{"meeting": meeting}).Decode(&stored)
		switch {
		case err == nil:
			e.live = stored
			e.persistedAt = time.Now()
		case err == mongo.ErrNoDocuments:
		default:
			// don't cache a failed load; retry next time
			fmt.Printf("live: loading %s failed: %v\n", meeting, err)
			return e
		}
	}
	liveEntries[meeting] = e
	return e
}

// getLive returns a copy of the live state of a meeting.
func getLive(meeting string) model.Live {
	liveMu.Lock()
	defer liveMu.Unlock()
	l := liveEntryLocked(meeting).live
	if l.Message != nil {
		m := *l.Message
		l.Message = &m
	}
	return l
}

// RecordLiveData notes that timing data for a meeting was received now.
func RecordLiveData(meeting string, source string) {
	now := time.Now()
	liveMu.Lock()
	e := liveEntryLocked(meeting)
	e.live.LastDataAt = now
	e.live.LastDataSource = source
	persist := now.Sub(e.persistedAt) >= livePersistInterval
	if persist {
		e.persistedAt = now
	}
	liveMu.Unlock()

	if persist {
		persistLive(meeting, bson.M{"$set": bson.M{"last_data_at": now, "last_data_source": source}})
	}
}

// SetLiveMessage sets the manual message of a meeting (text and mode must be validated).
func SetLiveMessage(meeting string, text string, mode model.LiveMessageMode, user string) error {
	msg := model.LiveMessage{Text: text, Mode: mode, UpdatedAt: time.Now(), UpdatedBy: user}
	liveMu.Lock()
	liveEntryLocked(meeting).live.Message = &msg
	liveMu.Unlock()
	return persistLive(meeting, bson.M{"$set": bson.M{"message": msg}})
}

// ClearLiveMessage removes the manual message of a meeting.
func ClearLiveMessage(meeting string) error {
	liveMu.Lock()
	liveEntryLocked(meeting).live.Message = nil
	liveMu.Unlock()
	return persistLive(meeting, bson.M{"$unset": bson.M{"message": ""}})
}

func persistLive(meeting string, update bson.M) error {
	if liveCollection == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := liveCollection.UpdateOne(ctx, bson.M{"meeting": meeting}, update, options.Update().SetUpsert(true))
	if err != nil {
		fmt.Printf("live: saving %s failed: %v\n", meeting, err)
	}
	return err
}
