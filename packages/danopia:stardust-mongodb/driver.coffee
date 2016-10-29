Stardust.Engines.MongoDB = class StardustMongoDBEngine
  constructor: (opts={}) ->
    @tableName = opts.tableName ? 'Stardust'
    @table = new Mongo.Collection @tableName

  start: (@multi) ->

  unwrap: (item, collection) ->
    collection ?= @multi.collections[item.type]
    doc =
      _id: item.id
      _version: item.version

    doc.Slug = item.slug

    doc._createdAt = item.createdAt
    doc._createdBy = item.createdBy
    doc._modifiedAt = item.modifiedAt
    doc._modifiedBy = item.modifiedBy

    for key, type of collection.schema when (obj = item.props[key])?
      doc[key] = switch type
        when String then ''+obj
        when Date then new Date +obj
        when Number then +obj
        when Boolean then !!obj

    return doc

  insertProps: (collection, props, {userId} = {}) ->
    random = DDP.randomStream('/collection/' + collection.name)
    _id = random.id()

    item =
      type: collection.name
      id: _id
      _id: [collection.name, _id].join ':'
      createdAt: new Date
      props: {}
      version: 1

    if userId
      item.createdBy = userId

    if collection.slug?
      item.slug = Stardust.slugify collection.slug.call(props).join('-')

    for key, type of collection.schema when props[key]?
      item.props[key] = switch type
        when String then ''+props[key]
        when Date then new Date(+props[key])
        when Number then +props[key]
        when Boolean then !!props[key]

    @table.insert item
    return @unwrap item, collection

StardustMongoDBEngine.QueryCursor = class StardustMongoDBQueryCursor
  constructor: (@coll, opts) ->
    @reactive = opts.reactive ? true
    @filter = opts.filter ? {}

    {@engine} = @coll.stardust

  # publication glue
  _getCollectionName: -> @coll.name
  _publishCursor: (sub) ->
    observeHandle = @observeChanges
      added: (id, fields) => sub.added @coll.name, id, fields
      changed: (id, fields) => sub.changed @coll.name, id, fields
      removed: (id) => sub.removed @coll.name, id
    sub.onStop -> observeHandle.stop()
    return observeHandle

  forEach: (callback, thisArg) -> # reactive, sequential, (doc, idx, @)
    console.log @coll.name, 'for each'
    @engine.table
      .find($and: [{type: @coll.name}, @filter])
      .map (item) => @engine.unwrap(item)
      .forEach callback, thisArg
  map: (callback, thisArg) -> # reactive, parallel, (doc, idx, @)
    console.log @coll.name, 'map'
    @engine.table
      .find($and: [{type: @coll.name}, @filter])
      .map (item) => @engine.unwrap(item)
      .map callback, thisArg
  fetch: -> # reactive, blocking
    console.log @coll.name, 'fetch'
    @engine.table
      .find($and: [{type: @coll.name}, @filter])
      .map (item) => @engine.unwrap(item)
  count: -> # reactive
    console.log @coll.name, 'count'
    @engine.table
      .find($and: [{type: @coll.name}, @filter])
      .count()

  observe: (cbs) -> # starts live query, blocks for initial results
    console.log @coll.name, 'observe'
    added: (doc) ->
    addedAt: (doc, atIdx, before) ->
    changed: (newDoc, oldDoc) ->
    changedAt: (newDoc, oldDoc, atIdx) ->
    removed: (oldDoc) ->
    removedAt: (oldDoc, atIdx) ->
    movedTo: (doc, fromIdx, toIdx, before) ->

    stop: => # auto runs when parent autorun stops, if any
      console.log @coll.name, 'observe STOP'

  observeChanges: (cbs) ->
    observation = @engine.table
      .find($and: [{type: @coll.name}, @filter])
      .observeChanges
        added: (_id, fields) =>
          [type, id] = _id.split ':'
          cbs.added? id, @engine.unwrap fields
        #addedBefore: (id, fields, before) =>
        #  [type, id] = _id.split ':'
        #  cbs.addedBefore? id, @unwrap(fields), before
        changed: (id, fields) =>
          [type, id] = _id.split ':'
          cbs.changed? id, @engine.unwrap(fields)
        movedBefore: (id, before) =>
          [type, id] = _id.split ':'
          cbs.movedBefore? id, before
        removed: (id) =>
          [type, id] = _id.split ':'
          cbs.removed? id

    stop: => # should run when parent autorun stops, if any?
      observation.stop()

Stardust.Engine = Stardust.Engines.MongoDB
