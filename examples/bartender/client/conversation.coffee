Session.set 'fragment id'
Session.set 'fragments', []

Meteor.startup -> Meteor.autorun ->
  Session.set 'player id', Characters.findOne(Slug: 'you')?._id

moveToFragment = (fragment) -> if fragment
  console.log 'moving to frag id', fragment._id,
      'from', Session.get('fragment id')
  Session.set 'fragment id', fragment._id
  Tracker.nonreactive ->
    # TODO: trim history
    Session.set 'fragments', Session.get('fragments').concat([fragment._id])

Template.Conversation.onCreated ->
  # Bot fragment choices
  @autorun ->
    nextFragments = Fragments.find
      Previous: Session.get 'fragment id'
      Character: $ne: Session.get('player id')

    console.log 'nextfrags', Session.get('fragment id'), nextFragments.fetch()
    if nextFragments.count()
      moveToFragment nextFragments.fetch()[0]

Template.Conversation.helpers
  fragment: ->
    Fragments.findOne Session.get('fragment id')

  # Fragment history
  fragments: ->
    Session.get('fragments').map (id) ->
      Fragments.findOne(id)

  # Player fragment choices
  nextFragments: ->
    Fragments.find
      Previous: Session.get('fragment id')
      Character: Session.get('player id')

  # Character for current fragment
  characterInfo: ->
    Characters.findOne @Character

Template.Conversation.events
  'click .select-fragment': (evt) ->
    evt.preventDefault()
    moveToFragment @
