package worker

import log "github.com/sirupsen/logrus"

func healthCheck(err error) {
	if err != nil {
		log.Errorf("unhealthy: %+v", err)
	}
}
