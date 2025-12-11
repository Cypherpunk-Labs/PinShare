package psfs

// create object called AllowedList with a name map to bool value
// keep this list in lower as using strings.ToLower() in evaluations.
var AllowedList = map[string]bool{
	// CAD filetypes
	"stl":    true,
	"gcode":  true,
	"goo":    true,
	"obj":    true,
	"sldprt": true,
	"sldasm": true,
	"dwg":    true,
	"dxf":    true,
	"f3d":    true,
	"f2d":    true,
	"fbx":    true,
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
	"pptx": true,
	// multimedia
	"avi": false,
	"mov": false,
	"mp4": true,
	"mp3": true,
	"png": true,
	"jpg": true,
	"svg": true,
	// plaintext
	"txt": true,
	"rtf": true,
	"md":  true,
}
