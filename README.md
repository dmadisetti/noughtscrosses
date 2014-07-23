Noughts and Crosses - Dumb AI
=========

[![Build Status](https://travis-ci.org/dmadisetti/noughtscrosses.png)](https://travis-ci.org/dmadisetti/noughtscrosses)

Hacked together for kicks.

This simple project plays tictactoe. Starts off very dumb and progressivly gets better. After game 150, seemed to deterministically draw. Cool stuff. I'm sure there's some fancy math involving minimum number of games to fatalistic dataset, but I'll leave that to the curious. Objectives complete!

[See the project in action](http://nought-crosses.appspot.com/)

Build yon Dart.
---
Make sure to `pub get` before you do anything. When you're ready to build, `dart2js the.dart -o the.js`

Building for yourself?
---
Check out `seed`. It's a [google-datastore backup](https://developers.google.com/appengine/docs/adminconsole/datastoreadmin#restoring_data_to_another_app) of `Records` that leaves the game with some intelligence.

Objectives:

- Go - `Check`
- Dart - `Check`
- GAE datastore - `Check`
- Basic AI - `Check`

Todo:

- Clean
- Learn from mistakes
- Document???

To-maybe-do:

- Prevent Dart interaction between requests.
- Style UI