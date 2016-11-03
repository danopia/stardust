dynamodb = new AWS.DynamoDB
dynamodbstreams = new AWS.DynamoDBStreams

Stardust.Engines.DynamoDB = class StardustDynamoDBEngine
  constructor: ({@tableName}) ->
    @tableInfo = dynamodb.describeTableSync
      TableName: @tableName
    .Table

    for key in @tableInfo.KeySchema
      @[key.KeyType.toLowerCase() + 'Key'] = key.AttributeName

    @watchedCursors = new Set

  start: (@multi) ->
    @streamSubs = []
    @startStream() if @tableInfo.LatestStreamArn

  startStream: ->
    # TODO: assert StreamViewType, StreamStatus, LastEvaluatedShardId
    {Shards} = dynamodbstreams.describeStreamSync
      StreamArn: @tableInfo.LatestStreamArn
      Limit: 100
    .StreamDescription

    for shard in Shards
      #unless shard.SequenceNumberRange.EndingSequenceNumber
      @startShard shard.ShardId

  startShard: (shardId) ->
    # TODO: wait for parents

    {ShardIterator} = dynamodbstreams.getShardIteratorSync
      ShardId: shardId
      ShardIteratorType: 'LATEST'
      StreamArn: @tableInfo.LatestStreamArn

    poll = => try
      {Records, NextShardIterator} = dynamodbstreams.getRecordsSync
        ShardIterator: ShardIterator
        Limit: 10

      for rec in Records
        # {eventID, eventName, dynamodb}
        # also eventVersion, eventSource, awsRegion
        {ApproximateCreationDateTime, Keys, NewImage} = rec.dynamodb
        #console.log 'processing', rec.eventName, Keys, NewImage
        @processUpdate NewImage

      if NextShardIterator
        ShardIterator = NextShardIterator
        Meteor.setTimeout poll, 1000
    catch err
      console.warn 'Shard poll', err.stack
      Meteor.setTimeout poll, 10000 if ShardIterator

    console.log 'Warming shard', shardId
    poll()

  unwrap: (item, collection) ->
    collection ?= @multi.collections[item[@hashKey].S]
    doc =
      _id: item[@rangeKey].S
      _version: +item.Version.N

    doc.Slug = item.Slug.S if item.Slug

    doc._createdAt = new Date item.CreatedAt.S if item.CreatedAt
    doc._createdBy = item.CreatedBy.S if item.CreatedBy
    doc._modifiedAt = new Date item.ModifiedAt.S if item.ModifiedAt
    doc._modifiedBy = item.ModifiedBy.S if item.ModifiedBy

    for key, type of collection.schema when obj = item['Doc.' + key]
      doc[key] = switch type
        when String then obj.S
        when Date then new Date obj.S
        when Number then +obj.N
        when Boolean then obj.BOOL

    return doc

  insertProps: (collection, props, {userId} = {}) ->
    random = DDP.randomStream('/collection/' + collection.name)
    props._id = random.id()

    item =
      CreatedAt: S: new Date().toISOString()
      Version: N: '1'

    if userId
      item.CreatedBy = S: userId

    item[@hashKey] = S: collection.name
    item[@rangeKey] = S: props._id

    if collection.slug?
      item.Slug = S: Stardust.slugify collection.slug.call(props).join('-')

    for key, type of collection.schema when props[key]?
      obj = switch type
        when String then S: ''+props[key]
        when Date then S: props[key].toISOString()
        when Number then N: ''+(+props[key])
        when Boolean then BOOL: !!props[key]

      unless obj.S is ''
        item['Doc.' + key] = obj

    dynamodb.putItemSync
      Item: item
      TableName: @tableName
      ConditionExpression: "attribute_not_exists(#{@rangeKey})"

    return @unwrap item, collection

  processUpdate: (newRecord) ->
    doc = @unwrap newRecord
    id = doc._id
    delete doc._id

    console.log 'hi mom', newRecord
    @watchedCursors.forEach (cursor) ->
      cursor._processUpdate 'added', id, doc

StardustDynamoDBEngine.QueryCursor = class StardustDynamoDBQueryCursor
  constructor: (@coll, opts) ->
    @reactive = opts.reactive ? true
    @filter = opts.filter ? {}

    @watches = new Set
    {@engine} = @coll.stardust

  _processUpdate: (type, id, doc) ->
    # TODO: actually match filter
    if Object.keys(@filter).length isnt 0
      return

    @watches.forEach (watcher) ->
      watcher.versions.set id, doc._version
      watcher.ids.add id
      watcher.cbs.added? id, doc
      #watcher.cbs.addedBefore? id, doc, null

  _subscribe: (watcher) ->
    @watches.add watcher
    @engine.watchedCursors.add @
    return @_unsub.bind(@, watcher)
  _unsub: (watcher) ->
    @watches.delete watcher
    if @watches.size is 0
      @engine.watchedCursors.delete @

  # publication glue
  _getCollectionName: -> @coll.name
  _publishCursor: (sub) ->
    # publications are never ordered
    observeHandle = @observeChanges
      added: (id, fields) => sub.added @coll.name, id, fields
      changed: (id, fields) => sub.changed @coll.name, id, fields
      removed: (id) => sub.removed @coll.name, id
    sub.onStop -> observeHandle.stop()
    return observeHandle

  # This is the actual blocking query implementation
  fetch: -> # reactive, blocking
    #console.log @coll.name, 'fetch'
    list = []
    {stop} = @observeChanges
      added: (id, doc) ->
        list.push doc
    stop()
    return list

  # Blocking helpers, currently implemented with @fetch()
  forEach: (callback, thisArg) -> # reactive, sequential, (doc, idx, @)
    #console.log @coll.name, 'for each'
    @fetch().forEach callback, thisArg
  map: (callback, thisArg) -> # reactive, parallel, (doc, idx, @)
    #console.log @coll.name, 'map'
    @fetch().map callback, thisArg
  count: -> # reactive
    #console.log @coll.name, 'count'
    @fetch().length

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
    #console.log @coll.name, 'observeChanges'

    watcher =
      versions: new Map # id => version
      ids: new Set      # id
      cbs: cbs

    {Items, LastEvaluatedKey} = dynamodb.querySync
      TableName: @engine.tableName
      # ProjectionExpression
      ConsistentRead: true
      # ExclusiveStartKey
      KeyConditionExpression: '#hashKey = :collName'
      ExpressionAttributeNames:
        '#hashKey': @engine.hashKey
      ExpressionAttributeValues:
        ':collName': S: @coll.name
      # TODO: actually filter
    # TODO: LastEvaluatedKey

    for item in Items
      doc = @engine.unwrap item, @coll
      id = doc._id
      delete doc._id

      watcher.versions.set id, doc._version
      watcher.ids.add id
      watcher.cbs.added? id, doc
      #watcher.cbs.addedBefore? id, doc, null

    stop: @_subscribe watcher

    ###
    added: (id, fields) ->
    addedBefore: (id, fields, before) ->
    changed: (id, fields) ->
    movedBefore: (id, before) ->
    removed: (id) ->
    ###

Stardust.Engine = Stardust.Engines.DynamoDB
