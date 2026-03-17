package domain

type Status string

const (
	StatusAdded     Status = "added"
	StatusRemoved   Status = "removed"
	StatusChanged   Status = "changed"
	StatusUnchanged Status = "unchanged"
)

type DiffField struct {
	Field  string
	Label  string
	Old    interface{}
	New    interface{}
	Status Status
}

type DiffResponse struct {
	RecordID  int64
	Action    string
	ChangedBy string
	ChangedAt string
	Fields    []DiffField
}
