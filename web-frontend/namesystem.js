const root = "/~~";

Vue.component('entry-item', {
  template: '#entry-item',
  props: {
    name: String,
    path: String,
    startOpen: Boolean,
  },
  data: function () {
    if (this.startOpen) {
      this.request()
    }

    return {
      entry: {},
      open: !!this.startOpen,
    };
  },
  computed: {
    isFolder: function () {
      return this.entry.type === "Folder";
    },
    icon: function () {
      switch (this.entry.type) {
        case "Folder":
          return this.open ? "folder_open" : "folder";
        default:
          return this.open ? "remove" : "add";
      }
    },
  },
  methods: {
    toggle: function () {
      this.open = !this.open;

      if (this.open && !this.requested) {
        this.request();
      }
    },
    request: function () {
      this.requested = true;

      fetch(this.path, {headers: {Accept: 'application/json'}})
        .then(x => x.json())
        .then(x => this.entry = x);
    },
    /*changeType: function () {
      if (!this.isFolder) {
        Vue.set(this.entry, 'children', [])
        this.addChild()
        this.open = true
      }
    },
    addChild: function () {
      this.model.children.push({
        name: 'new stuff'
      })
    }*/
  }
});

Vue.component('create-entry-item', {
  template: '#create-entry-item',
  props: {
    parent: String,
    startOpen: Boolean,
  },
  data: function () {
    return {
      type: null,
      open: !!this.startOpen,
    };
  },
  computed: {
  },
  methods: {
    toggle: function () {
      this.open = !this.open;
    },
    create: function () {
      this.requested = true;

      fetch(this.path, {headers: {Accept: 'application/json'}})
        .then(x => x.json())
        .then(x => this.entry = x);
    },
  }
});

var app = new Vue({
  el: '#app',
  data: {
    rootPath: "/~~/",
  },
});
