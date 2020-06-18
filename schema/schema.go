package schema

const (
	DocumentFileName  = "document.pdf"
	DocumentImageName = "image.jpg"
	TextFileName      = "text.txt"
	ImagesFolderName  = "images"
	TextFolderName    = "texts"
)

// Details represents the basic elements of page details.
type Details struct {
	Text string `json:"text"`
}

// SubmissionDetails represents the element details of a submission.
type SubmissionDetails struct {
	UUID     string    `json:"uuid"`
	FileName string    `json:"pdf_filename"`
	Path     string    `json:"path"`
	Pages    []Details `json:"page_details"`
}
