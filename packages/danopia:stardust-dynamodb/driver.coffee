`Stardust = {};`

dynamodb = new AWS.DynamoDB
dynamodbstreams = new AWS.DynamoDBStreams

slugify = (text) ->
  text.toString().toLowerCase()
    .split(':')[0]
    .replace(/\s+/g, '-')           # Replace spaces with -
    .replace(/[^\w\-]+/g, '')       # Remove all non-word chars
    .replace(/\-\-+/g, '-')         # Replace multiple - with single -
    .replace(/^-+/, '')             # Trim - from start of text
    .replace(/-+$/, '')             # Trim - from end of text

Stardust.Multi = class StardustMulti
  constructor: (opts={}) ->
    @tableName = opts.tableName ? 'Stardust'
    @collections = {}

    @tableInfo = dynamodb.describeTableSync
      TableName: @tableName
    .Table

    for key in @tableInfo.KeySchema
      @[key.KeyType.toLowerCase() + 'Key'] = key.AttributeName

    @streamSubs = []
    @startStream() if @tableInfo.LatestStreamArn

  collection: (name, opts) ->
    @collections[name] ?= new Stardust.Collection @, name, opts

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

      for record in Records
        @processRecord record

      if NextShardIterator
        ShardIterator = NextShardIterator
        Meteor.setTimeout poll, 1000
    catch err
      console.warn 'Shard poll', err.stack
      Meteor.setTimeout poll, 10000 if ShardIterator

    console.log 'Warming shard', shardId
    poll()

  processRecord: ({eventID, eventName, dynamodb}) ->
    # also eventVersion, eventSource, awsRegion
    {ApproximateCreationDateTime, Keys, NewImage} = dynamodb
    console.log 'processing', eventName, Keys, NewImage

    doc = People.unwrap NewImage
    id = doc._id
    delete doc._id

    #versions[id] = doc._version
    #ids.push id

    global.cbs.added? id, doc
    global.cbs.addedBefore? id, doc, null

Stardust.QueryCursor = class StardustQueryCursor
  constructor: (@coll, opts) ->
    @reactive = opts.reactive ? true
    @filter = opts.filter ? {}

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
  map: (callback, thisArg) -> # reactive, parallel, (doc, idx, @)
    console.log @coll.name, 'map'
  fetch: -> # reactive, blocking
    console.log @coll.name, 'fetch'
  count: -> # reactive
    console.log @coll.name, 'count'

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
    console.log @coll.name, 'observeChanges'

    {Items, LastEvaluatedKey} = dynamodb.querySync
      TableName: @coll.stardust.tableName
      # ProjectionExpression
      ConsistentRead: true
      # ExclusiveStartKey
      KeyConditionExpression: '#hashKey = :collName'
      ExpressionAttributeNames:
        '#hashKey': @coll.stardust.hashKey
      ExpressionAttributeValues:
        ':collName': S: @coll.name
      # TODO: actually filter
    # TODO: LastEvaluatedKey

    versions = {}
    ids = []

    for item in Items
      doc = @coll.unwrap item
      id = doc._id
      delete doc._id

      versions[id] = doc._version
      ids.push id

      global.cbs = cbs
      cbs.added? id, doc
      cbs.addedBefore? id, doc, null

    ###
    added: (id, fields) ->
    addedBefore: (id, fields, before) ->
    changed: (id, fields) ->
    movedBefore: (id, before) ->
    removed: (id) ->
    ###

    stop: => # should run when parent autorun stops, if any?
      console.log @coll.name, 'observeChanges STOP'
      # TODO

Stardust.Collection = class StardustCollection
  constructor: (@stardust, @name, opts) ->
    {@schema, @slug} = opts
    self = @

    Meteor.methods
      "/#{@name}/insert": (doc) ->
        random = DDP.randomStream('/collection/' + self.name)
        doc._id = random.id()

        item =
          CreatedAt: S: new Date().toISOString()
          Version: N: '1'

        if @userId
          item.CreatedBy = S: @userId

        item[self.stardust.hashKey] = S: self.name
        item[self.stardust.rangeKey] = S: doc._id

        if self.slug?
          item.Slug = S: slugify self.slug.call(doc).join('-')

        for key, type of self.schema when doc[key]?
          obj = switch type
            when String then S: ''+doc[key]
            when Date then S: doc[key].toISOString()
            when Number then N: ''+(+doc[key])
            when Boolean then BOOL: !!doc[key]

          unless obj.S is ''
            item['Doc.' + key] = obj

        dynamodb.putItemSync
          Item: item
          TableName: self.stardust.tableName
          ConditionExpression: "attribute_not_exists(#{self.stardust.rangeKey})"

        return self.unwrap item

  unwrap: (item) ->
    doc =
      _id: item[@stardust.rangeKey].S
      _version: +item.Version.N

    doc.Slug = item.Slug.S if item.Slug

    doc._createdAt = new Date item.CreatedAt.S if item.CreatedAt
    doc._createdBy = item.CreatedBy.S if item.CreatedBy
    doc._modifiedAt = new Date item.ModifiedAt.S if item.ModifiedAt
    doc._modifiedBy = item.ModifiedBy.S if item.ModifiedBy

    for key, type of @schema when obj = item['Doc.' + key]
      doc[key] = switch type
        when String then obj.S
        when Date then new Date obj.S
        when Number then +obj.N
        when Boolean then obj.BOOL

    return doc

  find: (filter, opts={}) ->
    return new Stardust.QueryCursor @, opts
