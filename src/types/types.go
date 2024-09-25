package types

type ApiResponse struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
	Data    User    `json:"data"`
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
