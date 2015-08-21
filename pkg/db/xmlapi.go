/*
Copyright © 2014–5 Brad Ackerman.

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

package db

func (d *dbInterface) APIKeys(userID int) ([]*XMLAPIKey, error) {
	rows, err := d.getAPIKeysStmt.Queryx(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]*XMLAPIKey, 0, 2)
	for rows.Next() {
		key := &XMLAPIKey{}
		err = rows.StructScan(key)
		if err != nil {
			return nil, err
		}
		results = append(results, key)
	}
	return results, nil
}

func (d *dbInterface) DeleteAPIKey(userID, keyID int) error {
	_, err := d.deleteAPIKeyStmt.Exec(userID, keyID)
	return err
}

func (d *dbInterface) AddAPIKey(key XMLAPIKey) error {
	_, err := d.addAPIKeyStmt.Exec(key.User, key.ID, key.VerificationCode, key.Description)
	return err
}
