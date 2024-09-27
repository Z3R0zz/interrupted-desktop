package types

type ApiResponse struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
	Data    User    `json:"data"`
}

type StatsResponse struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
	Data    Stats   `json:"data"`
}

type PasteResponse struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
	Data    Paste   `json:"data"`
}

type Paste struct {
	Url string `json:"url"`
}

type Stats struct {
	Uploads  int    `json:"uploads"`
	Pastes   int    `json:"pastes"`
	Storage  string `json:"storage"`
	UID      int    `json:"uid"`
	Joined   string `json:"joined_at"`
	Invitees int    `json:"invitees"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type FileResponse struct {
	Success bool   `json:"success"`
	IOS     string `json:"IOS"`
	Files   []File `json:"files"`
}

type File struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
}

type LoginResponse struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
	Data    struct {
		ApiKey string
	} `json:"data"`
}

type GalleryResponse struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
	Data    []ImageData `json:"data"`
}

type ImageData struct {
	URL string `json:"url"`
}
