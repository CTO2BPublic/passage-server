package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
)

var Config = config.GetConfig()
var Db = dbdriver.GetDriver()
