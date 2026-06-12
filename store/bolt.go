package store

import (
	"encoding/json"

	"github.com/korfra/veyra/model"
	bolt "go.etcd.io/bbolt"
)

var bucketRules = []byte("rules")
var bucketSettings = []byte("settings")

type Store struct {
	db *bolt.DB
}

func New(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketRules)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(bucketSettings)
		return err
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) Save(rule model.Rule) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRules)

		data, err := json.Marshal(rule)
		if err != nil {
			return err
		}

		return b.Put([]byte(rule.PublicPattern), data)
	})
}

func (s *Store) Get(publicPattern string) (*model.Rule, error) {
	var rule model.Rule

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRules)

		v := b.Get([]byte(publicPattern))
		if v == nil {
			return nil
		}

		return json.Unmarshal(v, &rule)
	})

	if err != nil {
		return nil, err
	}

	if rule.PublicPattern == "" {
		return nil, nil
	}

	return &rule, nil
}

func (s *Store) List() ([]model.Rule, error) {
	var rules []model.Rule

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRules)

		return b.ForEach(func(k, v []byte) error {
			var rule model.Rule
			if err := json.Unmarshal(v, &rule); err != nil {
				return err
			}
			rules = append(rules, rule)
			return nil
		})
	})

	return rules, err
}

func (s *Store) Delete(publicPattern string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRules)
		return b.Delete([]byte(publicPattern))
	})
}

func (s *Store) SetSetting(key, val string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSettings)
		return b.Put([]byte(key), []byte(val))
	})
}

func (s *Store) GetSetting(key string) (string, error) {
	var val string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSettings)
		v := b.Get([]byte(key))
		if v != nil {
			val = string(v)
		}
		return nil
	})
	return val, err
}
