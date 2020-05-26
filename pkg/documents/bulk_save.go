package documents

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

func key(doc CollectionEntry) string {
	name := strings.TrimSpace(doc.GetName())
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ". ", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func BulkSave(db *bolt.DB, collectionname string, docs []CollectionEntry) error {
	err := db.Batch(func(tx *bolt.Tx) error {
		masterCollectionBucket, txErr := tx.CreateBucketIfNotExists([]byte(collectionname))
		if txErr != nil {
			return errors.Wrapf(txErr, "error creating %s bucket", collectionname)
		}

		for _, doc := range docs {
			if doc.GetName() == "" {
				continue
			}
			bucketKey := key(doc)

			docBucket, txErr := tx.CreateBucketIfNotExists([]byte(bucketKey))
			if txErr != nil {
				return errors.Wrapf(txErr, "error creating %s bucket", bucketKey)
			}
			if txErr := masterCollectionBucket.Put([]byte(bucketKey), []byte(bucketKey)); txErr != nil {
				return errors.Wrapf(txErr, "error creating %s record in %s", bucketKey, collectionname)
			}
			docBody, txErr := json.Marshal(doc)
			if txErr != nil {
				return errors.Wrap(txErr, "JSON marshal error")
			}
			if txErr := docBucket.Put([]byte(doc.GetWhen().Format(time.RFC3339)), docBody); txErr != nil {
				return errors.Wrapf(txErr, "error creating %s record in %s", doc.GetName(), bucketKey)
			}
		}
		return nil
	})
	return err
}
