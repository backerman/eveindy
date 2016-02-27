# Copyright © 2014–6 Brad Ackerman.
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

# Plugins we use
bower = require 'gulp-bower'
coffee = require 'gulp-coffee'
coffeelint = require 'gulp-coffeelint'
concat = require 'gulp-concat'
del = require 'del'
fs = require 'fs'
gulp = require 'gulp'
gulpif = require 'gulp-if'
gutil = require 'gulp-util'
lazypipe = require 'lazypipe'
less = require 'gulp-less'
lessPluginAutoPrefix = require 'less-plugin-autoprefix'
merge2 = require 'merge2'
minifycss = require 'gulp-minify-css'
minifyhtml = require 'gulp-minify-html'
ngAnnotate = require 'gulp-ng-annotate'
notifier = require 'node-notifier'
notify = require 'gulp-notify'
request = require 'request'
shell = require 'gulp-shell'
spawn = require 'child_process'
  .spawn
template = require 'gulp-template'
uglify = require 'gulp-uglify'
useref = require 'gulp-useref'
sourcemaps = require 'gulp-sourcemaps'

config =
  bower: './bower_components'
  dest: './dist'
  temp: './temp'
  development: false
  goSrc: [
    './cmd/**'
    './pkg/**'
    ]
  assets: './assets'
  jsDest: './js/site.js'
  outpostURL: 'https://api.eveonline.com/eve/ConquerableStationList.xml.aspx'
  outpostFile: './conquerable-stations.xml'

# This isn't as awesome as let*, sadly.
do (path = (extn) ->
  config[extn + 'Src'] = config.assets + "/#{extn}/**/*." + extn ) ->
  path(e) for e in ['coffee', 'js', 'less']

# html isn't in its own directory.
config.htmlSrc = config.assets + '/**/*.html'

# images are copied over unaltered.
config.images = config.assets + '/image/**/*'

gulp.task 'gobuild', ->
  gulp.src ''
    .pipe shell 'go build ./cmd/...'
    .on 'error', notify.onError ->
      "Failed to build Go files."

gulp.task 'clean', ['cleango'], (cb) ->
  del [
    config.dest + '/**'
    config.temp + '/**'
    config.outpostFile
  ], cb

gulp.task 'cleango', ->
  gulp.src ''
    .pipe shell 'go clean ./...'
    .pipe shell 'rm -f server'

gulp.task 'less', ->
  autoprefix =
    new lessPluginAutoPrefix {browsers: 'last 2 versions'}
  gulp.src config.lessSrc
    .pipe sourcemaps.init()
    .pipe less {plugins: autoprefix}
    .pipe sourcemaps.write()
    .pipe gulp.dest config.temp+'/css'

gulp.task 'lint', ->
  gulp.src config.coffeeSrc
    .pipe coffeelint()
    .pipe coffeelint.reporter('coffeelint-stylish')
    .pipe coffeelint.reporter('fail')
    .on 'error', notify.onError (err) ->
      err.message

jsProd = lazypipe()
  .pipe ngAnnotate
  .pipe uglify

productionTasks = lazypipe()
  .pipe ->
    # FIXME: what about coffee?
    gulpif '*.js', jsProd()
  .pipe ->
    gulpif '*.css', minifycss()
  .pipe ->
    # FIXME: doesn't actually do this
    gulpif '*.html', minifyhtml {empty: true}

gulp.task 'scripts', ['lint'], ->
  merge2(gulp.src config.coffeeSrc
#    .pipe sourcemaps.init()
    .pipe coffee()
    .on 'error', notify.onError (err) ->
      err.message
#    .pipe sourcemaps.write()
    )
    .pipe concat config.jsDest
    .pipe gulp.dest config.dest

gulp.task 'html', ['less'], ->
  gulp.src config.htmlSrc
    # Parse template
    .pipe template {development: config.development}
    .on 'error', notify.onError (err) ->
      "Template error: #{err.message}"
    # Minify iff in production
    .pipe gulpif !config.development, productionTasks()
    .pipe useref()
    .pipe gulp.dest config.dest

gulp.task 'images', ->
  gulp.src config.images, {base: config.assets }
    .pipe gulp.dest config.dest

gulp.task 'set-development', ->
  config.development = true

gulp.task 'download-outposts', (done) ->
  # Download the list of sovnull outposts, if it's not already present on disk.
  # This task would use streams but I can't get it to work that way. :(
  fs.stat config.outpostFile, (err) ->
    if err
      console.log "Downloading outposts file."
      request config.outpostURL, (err, resp, body) ->
        if err
          throw err
        fs.writeFileSync config.outpostFile, body
        done()
    else
      console.log "Outposts file already exists; not redownloading."
      done()

gulp.task 'dev', [
  'set-development'
#  'download-outposts'
  'default'
]

gulp.task 'watch', ['dev'], ->
  # stolen from shorrockin/noted
  # https://github.com/shorrockin/noted/blob/master/gulp/tasks/watch.coffee
  server = null
  startServer = (signal) ->
    if server and signal
      server.kill signal
    # Bind only to localhost; use development mode
    process.env.EVEINDY_BIND = "[::1]:8000"
    process.env.EVEINDY_DEV = "true"
    server = spawn './server'
    server.stdout.on 'data', (data) ->
      process.stdout.write data
    server.stderr.on 'data', (data) ->
      process.stderr.write data
    server.on 'close', (code) ->
      notifier.notify
        title: "eveindy"
        sound: if code isnt 0 then "Basso" else false
        message: if code is 0 then "Restarting server." \
                  else "Server exited with code #{code}."

  gulp.watch config.goSrc, ['gobuild']
  gulp.watch './server', ->
    startServer 'SIGINT'
  gulp.watch config.htmlSrc, ['html']
  gulp.watch config.lessSrc, ['html']
  gulp.watch [
    config.coffeeSrc
    config.jsSrc
  ], ['scripts']

  startServer()

gulp.task 'default', ['html', 'scripts', 'images', 'gobuild']
