window.Contacts = new Meteor.Collection 'people'

Template.AddContact.events
  'submit form': (evt) ->
    evt.preventDefault()

    {FirstName, LastName, Email} = evt.target
    Contacts.insert
      FirstName: FirstName.value
      LastName: LastName.value
      Email: Email.value
    , (err, id) ->
      console.log err ? id
