package database

type DB struct {
	name string

	Sessions *sessionsCollection
}

func NewDB(url, username, password, name string) *DB {
	mongo := newMongo(url, username, password)

	return &DB{
		name:     name,
		Sessions: newSessionsCollection(mongo, name, "sessions"),
	}
}
