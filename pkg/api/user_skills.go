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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// SkillsHandler returns a web handler function that provides information on
// a toon's skills.
func SkillsHandler(localdb db.LocalDB, sess server.Sessionizer) web.HandlerFunc {
	handler := func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		userID := s.User
		charID, _ := strconv.Atoi(c.URLParams["charID"])
		skillGroupID, _ := strconv.Atoi(c.URLParams["skillGroupID"])
		skills, err := localdb.CharacterSkillGroup(userID, charID, skillGroupID)
		if err != nil {
			errorStr := "Unable to get character skills."
			http.Error(w, fmt.Sprintf(`{"status": "Error", "error": "%v"}`, errorStr),
				http.StatusInternalServerError)
			return
		}
		skillsJSON, _ := json.Marshal(&skills)
		w.Write(skillsJSON)
		return
	}

	return handler
}
