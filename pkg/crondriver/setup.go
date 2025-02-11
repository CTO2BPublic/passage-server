package crondriver

import "github.com/CTO2BPublic/passage-server/pkg/config"

type Cron struct {
	ApiToken string
}

var Driver = new(Cron)
var Config = config.GetConfig()

func GetDriver() *Cron {
	return Driver
}
