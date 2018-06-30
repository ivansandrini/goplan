package dao

import (
	"log"

	"github.com/ivansandrini/metrics/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MilestoneDetailsDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "milestone-details"
)

func (m *MilestoneDetailsDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

func (m *MilestoneDetailsDAO) FindAll() ([]models.MilestoneDetails, error) {
	var milestoneDetails []models.MilestoneDetails
	err := db.C(COLLECTION).Find(bson.M{}).All(&milestoneDetails)
	return milestoneDetails, err
}

func (m *MilestoneDetailsDAO) FindById(id string) (models.MilestoneDetails, error) {
	var movie models.MilestoneDetails
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&movie)
	return movie, err
}

func (m *MilestoneDetailsDAO) Insert(milestoneDetails models.MilestoneDetails) error {
	err := db.C(COLLECTION).Insert(&milestoneDetails)
	return err
}

func (m *MilestoneDetailsDAO) Delete(milestoneDetails models.MilestoneDetails) error {
	err := db.C(COLLECTION).Remove(&milestoneDetails)
	return err
}

func (m *MilestoneDetailsDAO) Update(milestoneDetails models.MilestoneDetails) error {
	err := db.C(COLLECTION).UpdateId(milestoneDetails.ID, &milestoneDetails)
	return err
}
