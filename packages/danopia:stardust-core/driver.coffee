`Stardust = {};`

Stardust.slugify = (text) ->
  text.toString().toLowerCase()
    .split(':')[0]
    .replace(/\s+/g, '-')           # Replace spaces with -
    .replace(/[^\w\-]+/g, '')       # Remove all non-word chars
    .replace(/\-\-+/g, '-')         # Replace multiple - with single -
    .replace(/^-+/, '')             # Trim - from start of text
    .replace(/-+$/, '')             # Trim - from end of text

Stardust.Engines = {}
Stardust.Engine = ->
  (alert ? console.log) 'No Stardust engine is loaded. Fire one up and try again.'

Stardust.Multi = class StardustMulti
  constructor: (opts={}) ->
    @tableName = opts.tableName ? 'Stardust'
    @collections = {}

    @engine = new Stardust.Engine
      tableName: @tableName
    @engine.start @

  collection: (name, opts) ->
    @collections[name] ?= new Stardust.Collection @, name, opts

  processUpdate: (newRecord) ->
    # TODO: use right collection
    doc = @engine.unwrap newRecord # TODO
    id = doc._id
    delete doc._id

    #versions[id] = doc._version
    #ids.push id

    # TODO: use right CB, copy API probably.
    global.cbs.added? id, doc
    global.cbs.addedBefore? id, doc, null

Stardust.Collection = class StardustCollection
  constructor: (@stardust, @name, opts) ->
    {@schema, @slug} = opts
    self = @

    # called by the clients
    Meteor.methods
      "/#{@name}/insert": (props) ->
        doc = self.stardust.engine.insertProps self, props, @
        return doc._id

  find: (filter, opts={}) ->
    return new @stardust.engine.constructor.QueryCursor @, opts

  # called on the server
  insert: (props) ->
    doc = @stardust.engine.insertProps @, props
    return doc._id
