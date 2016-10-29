Package.describe({
  summary: 'Plug-and-play DynamoDB for your Meteor collections, reactivity and all'
});

Package.onUse(function(api) {
  api.versionsFrom('1.4.1.2');
  api.use('coffeescript');
  api.use('random', 'server');

  api.mainModule('driver.coffee', 'server');
  // api.addFiles('util.coffee', 'server');
  api.export('Stardust', 'server');
});
