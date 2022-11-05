package main

//func (e *MongoDbOsm) GetAll() (users []dataformat.User, length int, err error) {
//	var cursor *mongo.Cursor
//
//	e.ClientUser = e.Client.Database(constants.KMongoDBDatabase).Collection(constants.KMongoDBCollectionUser)
//	cursor, err = e.ClientUser.Find(e.Ctx, bson.M{})
//	if err != nil {
//		return
//	}
//
//	err = cursor.All(e.Ctx, &users)
//	if err != nil {
//		return
//	}
//
//	length = len(users)
//	return
//}
