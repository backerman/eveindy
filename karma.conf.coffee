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

module.exports = (config) ->
  config.set
    basePath: ''
    frameworks: ['jasmine', 'fixture']
    preprocessors:
      '**/*.coffee': 'coffee'
      '**/*.json': 'json_fixtures'
      '**/*.html': 'ng-html2js'
    jsonFixturesPreprocessor:
      variableName: '__json__'
    files: [
      'bower_components/jquery/dist/jquery.js'
      'bower_components/angular/angular.js'
      'bower_components/datatables/media/js/jquery.dataTables.js'
      'bower_components/datatables/media/js/dataTables.bootstrap.js'
      'bower_components/angular-route/angular-route.js'
      'bower_components/angular-bootstrap/ui-bootstrap-tpls.js'
      'bower_components/angular-mocks/angular-mocks.js'
      'bower_components/angular-datatables/dist/angular-datatables.js'
      'bower_components/angular-datatables/dist/plugins/bootstrap/angular-datatables.bootstrap.js'
      'bower_components/angular-promise-buttons/dist/angular-promise-buttons.js'
      'assets/coffee/**/*.coffee'
      'spec/*.coffee'
      'spec/fixtures/*.json'
      'assets/view/directives/*.html'
    ]

    ngHtml2JsPreprocessor:
      stripPrefix: 'assets/'
      moduleName: 'directives_test'

    exclude: []
    port: 8080
    logLevel: config.LOG_INFO
    autoWatch: true
    browsers: ['Chrome']
    singleRun: false
