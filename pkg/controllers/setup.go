package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
	"github.com/CTO2BPublic/passage-server/pkg/eventdriver"
)

var Config = config.GetConfig()
var Db = dbdriver.GetDriver()
var Event = eventdriver.GetDriver()
