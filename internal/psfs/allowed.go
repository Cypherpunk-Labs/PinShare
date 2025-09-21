package psfs

// create object called AllowedList with a name map to bool value
// keep this list in lower as using strings.ToLower() in evaluations.
var AllowedList = map[string]bool{
	// CAD filetypes
	"stl":    true,
	"sldprt": true,
	"sldasm": true,
	"dwg":    true,
	"dxf":    true,
	"3dm":    true,
	"stp":    true,
	"step":   true,
	"igs":    true,
	"iges":   true,
	"3ds":    true,
	"sat":    true,
	// docs
	"pdf":  true,
	"doc":  false,
	"docx": false,
	// multimedia
	"avi": false,
	"mov": false,
	"mp4": false,
}
