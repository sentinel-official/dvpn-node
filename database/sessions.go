package database

import (
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

func (s sessionsCollection) AddSessionBandwidthSign(id string, bandwidthSign *BandwidthSign) error {
	selector := bson.M{
		"sessionID": id,
	}

	update := bson.M{
		"$push": bson.M{
			"bandwidthSigns": bandwidthSign,
		},
	}

	return s.mongo.UpdateOne(s.database, s.name, selector, update)
}

func (s sessionsCollection) GetSessionBandwidthSign(id string, index int64) (*BandwidthSign, error) {
	query := bson.M{
		"sessionID":            id,
		"bandwidthSigns.index": index,
	}

	selectors := bson.M{}

	var bandwidthSign BandwidthSign
	if err := s.mongo.GetOne(s.database, s.name, query, selectors, &bandwidthSign); err != nil {
		return nil, err
	}

	return &bandwidthSign, nil
}

func (s sessionsCollection) AddSessionBandwidthClientSign(id string, index int64, sign string) error {
	selector := bson.M{
		"sessionID":            id,
		"bandwidthSigns.index": index,
	}

	update := bson.M{
		"$set": bson.M{
			"bandwidthSigns.$.clientSign": sign,
		},
	}

	return s.mongo.UpdateOne(s.database, s.name, selector, update)
}
