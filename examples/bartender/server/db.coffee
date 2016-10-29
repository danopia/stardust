global.stardust = new Stardust.Multi

global.Characters = stardust.collection 'characters',
  schema:
    Name: String
    Gender: String
    Image: String
    Age: Number
  slug: -> [@Name]

global.Fragments = stardust.collection 'fragments',
  schema:
    Previous: String
    Character: String
    Text: String

global.Migrations = stardust.collection 'migrations',
  schema:
    Ordering: Number

Meteor.publish null, -> [
  Characters.find()
  Fragments.find()
]


Meteor.startup ->
  currentVersion = 0
  Migrations.find().forEach (migration) ->
    if currentVersion < migration.Ordering
      currentVersion = migration.Ordering

  if currentVersion < 1
    console.log 'Migrating to 1'
    
    mei = Characters.insert
      Name: 'Mei Chang'
      Gender: 'female'
      Image: 'http://st.depositphotos.com/1446089/4722/i/950/depositphotos_47220023-Nerd-asian-college-girl-with.jpg'
      Age: 21

    player = Characters.insert
      Name: 'You'
      Gender: 'male'
      Image: 'http://thumb1.shutterstock.com/display_pic_with_logo/2139845/415276084/stock-photo-bartender-pouring-a-cocktail-into-glass-close-up-no-face-415276084.jpg'
      Age: 34

    c1 = Fragments.insert
      Previous: null
      Character: mei
      Text: 'Hey, excuse me.'

    c2 = Fragments.insert
      Previous: c1
      Character: player
      Text: 'What can I do you for?'

    c3 = Fragments.insert
      Previous: c2
      Character: mei
      Text: 'Um...just a...um, tequila sunrise?'

    c4 = Fragments.insert
      Previous: c3
      Character: player
      Text: 'Coming right up.'

    c5 = Fragments.insert
      Previous: c4
      Character: player
      Text: 'You a university student? A lot of them in come in during exam season.'

    c6 = Fragments.insert
      Previous: c5
      Character: mei
      Text: 'Y...yeah.'

    c7 = Fragments.insert
      Previous: c6
      Character: mei
      Text: 'But I’m only here because my roommates dragged me out.'

    c8a = Fragments.insert
      Previous: c7
      Character: player
      Text: 'Why?'

    c9a = Fragments.insert
      Previous: c8a
      Character: mei
      Text: 'It’s...a really long story.'

    c8b = Fragments.insert
      Previous: c7
      Character: player
      Text: 'You can’t let people drag you around like that.'

    c9b = Fragments.insert
      Previous: c8b
      Character: mei
      Text: 'I guess, but in a way it’s for my own good.'

    c10b = Fragments.insert
      Previous: c9b
      Character: mei
      Text: 'I usually just go to class and then come back to my dorm and stay in.'

    c11b = Fragments.insert
      Previous: c10b
      Character: mei
      Text: 'Life’s easier that way.'

    Migrations.insert
      Ordering: 1
