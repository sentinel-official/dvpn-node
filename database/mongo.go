package database

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongo struct {
	url      string
	username string
	password string
}

func newMongo(url, username, password string) *mongo {
	return &mongo{
		url:      url,
		username: username,
		password: password,
	}
}

func (m *mongo) newSession() (*mgo.Session, error) {
	info := &mgo.DialInfo{
		Addrs:    []string{m.url},
		Timeout:  60 * time.Second,
		Username: m.username,
		Password: m.password,
	}

	return mgo.DialWithInfo(info)
}

func (m mongo) Insert(database, collection string, objects ...bson.M) error {
	session, err := m.newSession()
	if err != nil {
		return err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.Insert(objects)
}

func (m mongo) UpdateOne(database, collection string, selector, update bson.M) error {
	session, err := m.newSession()
	if err != nil {
		return err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.Update(session, update)
}

func (m mongo) UpdateMany(database, collection string, selector, update bson.M) (*mgo.ChangeInfo, error) {
	session, err := m.newSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.UpdateAll(session, update)
}

func (m mongo) GetOne(database, collection string, query, selectors bson.M, object interface{}) error {
	session, err := m.newSession()
	if err != nil {
		return err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.Find(query).Select(selectors).One(object)
}

func (m mongo) GetMany(database, collection string, query, selectors bson.M, object interface{}) error {
	session, err := m.newSession()
	if err != nil {
		return err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.Find(query).Select(selectors).All(object)
}

func (m mongo) Remove(database, collection string, selector bson.M) error {
	session, err := m.newSession()
	if err != nil {
		return err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.Remove(selector)
}

func (m mongo) RemoveAll(database, collection string, selector bson.M) (*mgo.ChangeInfo, error) {
	session, err := m.newSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()
	c := session.DB(database).C(collection)

	return c.RemoveAll(selector)
}
