package database

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

func (s sessionsCollection) AddSession(session Session) error {
	return nil
}
