global.stardust = new Stardust.Multi

global.Contacts = stardust.collection 'contacts',
  schema:
    FirstName: String
    LastName: String
    #Gender: String
    Phone: String
    Email: String
    Address: String
  slug: -> [@FirstName, @LastName]

Meteor.publish null, ->
  Contacts.find()
