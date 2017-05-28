const root = "/~~";

Vue.component('entry-item', {
  template: '#entry-item',
  props: {
    name: String,
    path: String,
    startOpen: Boolean
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

var app = new Vue({
  el: '#app',
  data: {
    rootPath: "/~~/",
  },
});
