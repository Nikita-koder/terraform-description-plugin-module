package schema

type Schema struct {
	Description         string
	MarkdownDescription string
	Attributes          map[string]Attribute
}

type Attribute interface {
	isAttribute()
	GetDescription() string
}

type StringAttribute struct {
	Description         string
	MarkdownDescription string
	Optional            bool
	Computed            bool
}

func (StringAttribute) isAttribute() {}
func (a StringAttribute) GetDescription() string {
	return a.Description
}

type SingleNestedAttribute struct {
	Description         string
	MarkdownDescription string
	Attributes          map[string]Attribute
	Optional            bool
}

func (SingleNestedAttribute) isAttribute() {}
func (a SingleNestedAttribute) GetDescription() string {
	return a.Description
}
