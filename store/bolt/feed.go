package bolt

import (
	"encoding/json"

	"github.com/mattmeyers/feedsync/store"
	"go.etcd.io/bbolt"
)

type FeedStore struct {
	db *bbolt.DB
}

func NewFeedStore(db *bbolt.DB) *FeedStore {
	return &FeedStore{db: db}
}

func (s *FeedStore) List() ([]store.Feed, error) {
	feeds := []store.Feed{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketFeeds))
		b.ForEach(func(k, v []byte) error {
			var f store.Feed
			err := json.Unmarshal(v, &f)
			if err != nil {
				return err
			}

			f.ID = btoi(k)

			feeds = append(feeds, f)
			return nil
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (s *FeedStore) Insert(f store.Feed) (uint, error) {
	var id uint64

	err := s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(f)
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte(bucketFeeds))
		id, _ = b.NextSequence()

		err = b.Put(itob(uint(id)), data)
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return uint(id), nil
}
