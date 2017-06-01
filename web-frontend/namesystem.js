const root = "/~~";

// TODO
window.require = function (names) {
  console.log("'Requiring'", names)
}

Vue.component('entry-item', {
  template: '#entry-item',
  props: {
    name: String,
    type: String,
    path: String,
    startOpen: Boolean,
  },
  data: function () {
    return {
      entry: {},
      open: !!this.startOpen,
      loader: this.startOpen ? this.load() : null,
    };
  },
  computed: {
    isFolder: function () {
      return this.type === "Folder";
    },
    icon: function () {
      switch (this.type) {
        case "Folder":
          return this.open ? "folder_open" : "folder";
        case undefined: // TODO: unugly
          return this.open ? "expand_less" : "chevron_right";
        default:
          return "insert_drive_file";
      }
    },
  },
  methods: {
    activate: function () {
      if (this.type === 'Folder') {
        this.open = !this.open;
        this.load();
      } else {
        app.openEditor({
          type: 'edit-' + this.type.toLowerCase(),
          label: this.name,
          icon: 'edit',
          path: this.path,
        });
      }
    },
    load: function () {
      if (!this.loader) {
        this.loader = fetch(root + this.path, {
          headers: {Accept: 'application/json'},
        }).then(x => x.json())
          .then(x => this.entry = x);
      }
      return this.loader;
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
    parentName: String,
  },
  methods: {
    activate: function () {
      app.openEditor({
        type: 'create-name',
        label: 'create (' + this.parentName + ')',
        icon: 'add',
        path: this.parent,
      });
    },
    /*create: function () {
      this.requested = true;

      fetch(this.path, {headers: {Accept: 'application/json'}})
        .then(x => x.json())
        .then(x => this.entry = x);
    },*/
  }
});

Vue.component('create-name', {
  template: '#create-name',
  props: {
    tab: Object,
  },
  data: function () {
    return {
    };
  },
  computed: {
  },
  methods: {
    activate: function () {
    },
  }
});

Vue.component('edit-file', {
  template: '#edit-file',
  props: {
    tab: Object,
  },
  data() {
    const pathParts = this.tab.path.split('/');
    return {
      source: '',
      editorOptions: {
        tabSize: 2,
        mode: {
          filename: pathParts[pathParts.length - 1],
        },
        styleActiveLine: true,
        lineWrapping: true,
        lineNumbers: true,
        line: true,
        styleSelectedText: true,
        matchBrackets: true,
        showCursorWhenSelecting: true,
        theme: "tomorrow-night-bright",
        extraKeys: { "Ctrl": "autocomplete" },
      }
    };
  },
  computed: {
  },
  methods: {
    activate: function () {
    },
    save() {
      fetch(root + this.tab.path, {
        method: 'PUT',
        body: this.source,
        headers: {
          'X-SD-Entry-Type': 'File',
          'Content-Type': 'text/plain',
        },
      }).then(x => x.text())
        .then(x => alert('File Save Response:\n\n' + x));
    },
  },
  created() {
    fetch(root + this.tab.path, {
      headers: {
        Accept: 'text/plain',
      },
    }).then(x => x.text())
      .then(x => this.source = x);
  },
});

var app = new Vue({
  el: '#app',
  data: {
    rootPath: "/~~",
    tabList: [],
    tabKeys: {},
    currentTab: null,
  },
  created() {
    //this.openEditor({"type":"edit-string","label":"init","icon":"edit","path":"/boot/init"});
    this.openEditor({"type":"edit-file","label":"namesystem.html","icon":"edit","path":"/n/osfs/namesystem.html"});
    //this.openEditor({"type":"create-name","label":"create (tmp)","icon":"add","path":"/tmp"});

    window.addEventListener('keydown', this.handleKeyDown);
  },
  destroyed () {
    window.removeEventListener('keydown', this.handleKeyDown);
  },
  methods: {
    // Focus or open a new editor for given details
    openEditor(deets) {
      deets.key = [deets.path, deets.type].join(':');
      if (deets.key in this.tabKeys) {
        this.activateTab(this.tabKeys[deets.key]);
      } else {
        console.log("Opening editor", deets.key, 'labelled', deets.label);
        this.currentTab = deets;
        this.tabList.push(deets);
        this.tabKeys[deets.key] = deets;
      }
    },
    activateTab(tab) {
      console.log("Switching to tab", tab.label);
      this.currentTab = tab;
    },
    closeTab(tab) {
      const idx = this.tabList.indexOf(tab);
      console.log("Closing tab", tab.label, "idx", idx);
      if (idx !== -1) {
        this.tabList.splice(idx, 1);
      }
      delete this.tabKeys[tab.key];

      if (this.currentTab === tab) {
        this.currentTab = this.tabList[0];
        const idx = this.tabList.indexOf(this.currentTab);
      }
    },

    handleKeyDown(evt) {
      const tab = (this.$refs.tabElems || [])
        .find(elem => elem.tab.key === this.currentTab.key);

      switch (true) {

      case evt.code === 'KeyS' && (evt.metaKey || evt.ctrlKey):
        if (tab) {
          evt.preventDefault();
          console.log('Saving tab:', tab.tab.label);
          tab.save();
        }
        break;

      case evt.code === 'KeyN' && evt.metaKey:
        if (tab) {
          evt.preventDefault();
          const pathParts = tab.tab.path.slice(1).split('/');
          this.openEditor({
            type: "create-name",
            label: "create (" + pathParts[pathParts.length - 2] + ")",
            icon: "add",
            path: "/" + pathParts.slice(0, -1).join('/'),
          });
        }
        break;

      }
    },
  },

});
