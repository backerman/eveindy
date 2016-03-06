/*
Copyright © 2014–6 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/backerman/eveindy/pkg/db"
	"github.com/robfig/cron"
)

// StartJobs starts the background jobs to update universe information.
func StartJobs(localdb db.LocalDB) {
	jobs := []struct {
		cronSpec string
		job      func()
	}{
		{"@every 1h", func() { updateOutposts(localdb) }},
	}
	c := cron.New()
	for _, j := range jobs {
		err := c.AddFunc(j.cronSpec, j.job)
		if err != nil {
			log.Fatalf("Unable to add background job: %v", err)
		}
		// Execute job on launch as well.
		go j.job()
	}
	c.Start()
}

// Periodic updates go here.

// updateOutposts grabs the outpost information and inserts it into the
// database.
func updateOutposts(localdb db.LocalDB) {
	log.Printf("Starting outposts update")
	start := time.Now()
	err := localdb.RepopulateOutposts()
	if err != nil {
		log.Printf("Error updating outposts: %v", err)
	} else {
		duration := time.Now().Sub(start)
		log.Printf("Finished outpost update in %.0f ms", duration.Seconds()*1000.0)
	}
}
