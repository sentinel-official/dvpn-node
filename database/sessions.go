package database

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type sessionsCollection struct {
	mongo    *mongo
	database string
	name     string
}

func newSessionsCollection(mongo *mongo, database, name string) *sessionsCollection {
	return &sessionsCollection{
		mongo:    mongo,
		database: database,
		name:     name,
	}
}

func (s sessionsCollection) AddSession(session *Session) error {
	return s.mongo.Insert(s.database, s.name, session)
}

func (s sessionsCollection) AddBandwidthSign(sessionID string, bandwidthSign *BandwidthSign) error {
	selector := bson.M{
		"sessionID": sessionID,
	}
	update := bson.M{
		"$push": bson.M{
			"bandwidthSigns": bandwidthSign,
		},
	}

	return s.mongo.UpdateOne(s.database, s.name, selector, update)
}

func (s sessionsCollection) GetBandwidthSign(sessionID, id string) (*BandwidthSign, error) {
	query := bson.M{
		"sessionID": sessionID,
	}
	selectors := bson.M{
		"_id": false,
		"bandwidthSigns": bson.M{
			"$elemMatch": bson.M{
				"_id": bson.ObjectIdHex(id),
			},
		},
	}

	var result []BandwidthSign
	if err := s.mongo.GetOne(s.database, s.name, query, selectors, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result[0], nil
}

func (s sessionsCollection) AddBandwidthClientSign(sessionID, id, sign string) error {
	selector := bson.M{
		"sessionID":          sessionID,
		"bandwidthSigns._id": bson.ObjectIdHex(id),
	}
	update := bson.M{
		"$set": bson.M{
			"bandwidthSigns.$.clientSign": sign,
			"bandwidthSigns.$.updatedAt":  time.Now().UTC(),
		},
	}

	return s.mongo.UpdateOne(s.database, s.name, selector, update)
}
