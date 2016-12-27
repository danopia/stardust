# TODO: publish dependencies too
# use reywood:publish-composite
# https://github.com/englue/meteor-publish-composite

Meteor.publish '/app-runtime', (packageId) ->
  check packageId, String

  # TODO: this is effectively auto-publish
  #       just scoped to the current app
  [
    DB.App.find _id: packageId
    DB.Resource.find {packageId}
    DB.Record.find {packageId}
  ]
