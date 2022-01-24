package bolt

import (
	"encoding/binary"

	"go.etcd.io/bbolt"
)

type bucketName string

const (
	bucketFeeds bucketName = "feeds"
)

func Open(dbPath string) (*bbolt.DB, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bbolt.Tx) error {
		for _, b := range []bucketName{bucketFeeds} {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return db, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi converts an 8-byte big endian representation to an int.
func btoi(v []byte) uint {
	return uint(binary.BigEndian.Uint64(v))
}
