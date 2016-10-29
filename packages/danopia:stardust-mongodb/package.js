Package.describe({
  summary: 'Plug-and-play MongoDB for your Meteor collections, reactivity and all'
});

Package.onUse(function(api) {
  api.versionsFrom('1.4.1.2');
  api.use('coffeescript');
  api.use('mongo', 'server');
  api.use('random', 'server');

  api.use('danopia:stardust-core');
  api.imply('danopia:stardust-core');
  api.mainModule('driver.coffee', 'server');
  // api.addFiles('util.coffee', 'server');
});
