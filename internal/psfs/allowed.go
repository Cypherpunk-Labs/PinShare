package psfs

// create object called AllowedList with a name map to bool value
// keep this list in lower as using strings.ToLower() in evaluations.
var AllowedList = map[string]bool{
	"pdf":  true,
	"doc":  false,
	"docx": false,
	"avi":  false,
	"mov":  false,
	"mp4":  false,
}
