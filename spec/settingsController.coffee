# Copyright © 2014–5 Brad Ackerman.
#
# Licensed under the Apache License, Version 2.0  the "License";
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

describe 'Controller: SettingsCtrl', () ->
  beforeEach () -> module 'eveindy'
  ctrl = undefined
  serverService = undefined
  apiKeys = undefined
  scope = undefined

  beforeEach () ->
    inject ($controller, Server, $rootScope) ->
      scope = $rootScope.$new()
      apiKeys = fixture.load('apiKeys.json')
      spyOn Server, 'apiForUser'
        .and.returnValue
          then: (callback) ->
            response =
              data: JSON.parse JSON.stringify apiKeys
            callback response
      spyOn Server, 'deleteApiKey'
        .and.callFake (keyid) ->
          then: (callback) ->
            callback
              status: 'OK'

      spyOn Server, 'addApiKey'
        .and.callFake (key) ->
          then: (callback) ->
            callback
              status: 'OK'

      serverService = Server
      ctrl = $controller 'SettingsCtrl',
        $scope: scope
      scope.$broadcast('login-status', true)

  it 'should get a user\'s API keys', (done) ->
    expect(serverService.apiForUser).toHaveBeenCalled()
    expect(serverService.apiForUser.calls.count()).toEqual 1
    expect((k.id for k in ctrl.apikeys).sort()).toEqual [123456, 234567, 345678]
    done()

  it 'should remove a deleted key from its local model', (done) ->
    ctrl.deleteKey 234567
    expect(k.id for k in ctrl.apikeys).not.toContain 234567
    expect(ctrl.apikeys.length).toEqual 2
    done()

  it 'should add keys to the user\'s account', (done) ->
    ctrl.newkey =
      id: 666
      userid: 101
      vcode: "abcdefg"
      label: "hijklmnop"
    ctrl.addKey()
    expect(ctrl.apikeys.length).toEqual 4
    expect(k.id for k in ctrl.apikeys).toContain 666
    done()
