Template.Contacts.helpers
  contacts: ->
    Contacts.find {},
      sort: LastName: 1
