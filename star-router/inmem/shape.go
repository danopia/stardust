package inmem

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
)

// Compiles a Shape out of a Folder
type Shape struct {
	cfg base.Folder
	validateFunc base.Function
	typeName string
	props []*Shape
	optional bool
}

var _ base.Shape = (*Shape)(nil)

func NewShape(config base.Folder) *Shape {
	// TODO: freeze config
	s := &Shape{cfg: config}

	s.typeName, _ = helpers.GetChildString(config, "type")
	optional, _ := helpers.GetChildString(config, "optional")
	s.optional = optional == "yes"

	propsEntry, ok := config.Fetch("props")
	if ok {
		propsFolder := propsEntry.(base.Folder)
		propNames := propsFolder.Children()
		s.props = make([]*Shape, 0, len(propNames))
		for _, propName := range propNames {
			if prop, ok := propsFolder.Fetch(propName); ok {
				var propCfg base.Folder
				switch prop := prop.(type) {

				case base.String:
					propCfg = NewFolderOf(prop.Name(),
						NewString("type", prop.Get()))

				case base.Folder:
					propCfg = prop

				default:
					log.Println("got unknown prop", prop, "for", s)
				}

				if propCfg != nil {
					s.props = append(s.props, NewShape(propCfg))
				}
			}
		}
	}

	s.validateFunc = NewFunction("validate", func(input base.Entry) (output base.Entry) {
		if s.Check(input) {
			output = NewString("result", "ok")
		}
		return
	})
	return s
}

func (e *Shape) Check(entry base.Entry) (ok bool) {
	if e.optional && entry == nil {
		return true
	}

	switch e.typeName {

	case "String":
		_, ok = entry.(base.String)

	case "Folder":
		if folder, castOk := entry.(base.Folder); castOk {
			ok = true
			for _, prop := range e.props {
				actual, getOk := folder.Fetch(prop.Name())
				if getOk {
					ok = prop.Check(actual)
				}
			}
		}

	}

	if !ok {
		log.Println("Validating failed:", entry, "against", e)
	}

	return
}

func (e *Shape) Name() string {
	return e.cfg.Name()
}

func (e *Shape) Children() []string {
	names := e.cfg.Children()
	names = append(names, "validate")
	return names
}

func (e *Shape) Fetch(name string) (entry base.Entry, ok bool) {
	if name == "validate" {
		return e.validateFunc, true
	}

	entry, ok = e.cfg.Fetch(name)
	return
}

func (e *Shape) Put(name string, entry base.Entry) (ok bool) {
	return false
}
