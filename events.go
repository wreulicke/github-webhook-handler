package webhook

type InstallationPart struct {
	Installation *Installation `json:"installation"`
}

type Installation struct {
	Id int64 `json:"id"`
}
